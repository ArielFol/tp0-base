package common

import (
	"bufio"
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"
	"strconv"
	"io"
	"encoding/csv"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchMaxAmount int
	RetriesMaxAmount int
	RetriesDelay     time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func sendAll(conn net.Conn, data []byte) error {
	totalSent := 0
	for totalSent < len(data) {
		n, err := conn.Write(data[totalSent:])
		if err != nil {
			return err
		}
		totalSent += n
	}
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM)

	file, err := os.Open("/data/agency.csv")
	if err != nil {
		log.Errorf("action: open_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	betReader := &BetReader{
		reader: reader,
	}
	

	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++{
		select {
		case <-sigChannel:
			log.Infof("action: shutdown | result: in_progress | client_id: %v", c.config.ID)
			if c.conn != nil {
				c.conn.Close()
				log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
			}
			return
		default:
		}

		id, err := strconv.ParseUint(c.config.ID, 10, 32)
		if err != nil {
			log.Errorf("invalid client id: %v", err)
			return
		}

		bets, err := readNextBets(betReader, c.config.BatchMaxAmount, uint32(id))
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("action: read_bets | result: fail | client_id: %v | error: %v", id, err)
			return
		}
		log.Infof("action: read_bets | result: success | client_id: %v | amount: %v", id, len(bets))

		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()
		if c.conn == nil {
			log.Errorf("action: connect | result: fail | client_id: %v", id)
			time.Sleep(c.config.LoopPeriod)
			continue
		}
		log.Infof("action: connect | result: success | client_id: %v", id)


		encodedBets, err := encodeBets(bets)
		if err != nil {
			log.Errorf("action: encode_bets | result: fail | client_id: %v | error: %v", id, err)
			return
		}
		log.Infof("action: encode_bets | result: success | client_id: %v | amount: %v", id, len(bets))


		if err := sendAll(c.conn, encodedBets); err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", id, err)
			return
		}
		log.Infof("action: send_message | result: success | client_id: %v | amount: %v", id, len(bets))

		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()
		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				id,
				err,
			)
			return
		}
		
		if msg != "200\n" {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: unexpected response '%v'",
				id,
				msg,
			)
			return
		}

		log.Infof("action: apuesta_enviada| result: success | cantidad: %v",
			len(bets),
		)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)
	}
	
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) NotifyFinish() {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM)

	select {
		case <-sigChannel:
			log.Infof("action: shutdown | result: in_progress | client_id: %v", c.config.ID)
			if c.conn != nil {
				c.conn.Close()
				log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
			}
			return
		default:
	}

	c.createClientSocket()
	if c.conn == nil {
		log.Errorf("action: connect | result: fail | client_id: %v", c.config.ID)
		return
	}
	log.Infof("action: connect | result: success | client_id: %v", c.config.ID)

	id, err := strconv.ParseUint(c.config.ID, 10, 32)
	if err != nil {
		log.Errorf("invalid client id: %v", err)
		return
	}

	encodedMsg, err := encodeFinishMessage(uint32(id))
	if err != nil {
		log.Errorf("action: encode_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	if err := sendAll(c.conn, encodedMsg); err != nil {
		log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	log.Infof("action: send_message | result: success | client_id: %v", c.config.ID)

	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	c.conn.Close()
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			id,
			err,
		)
		return
	}
		
	if msg != "200\n" {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: unexpected response '%v'",
			id,
			msg,
		)
		return
	}

	log.Infof("action: finalizacion enviada| result: success")
}

func (c *Client) GetResults() {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM)

	for attempts := 0; attempts < c.config.RetriesMaxAmount; attempts++ {
		select {
			case <-sigChannel:
				log.Infof("action: shutdown | result: in_progress | client_id: %v", c.config.ID)
				if c.conn != nil {
					c.conn.Close()
				log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
				}
				return
			default:
		}

		id, err := strconv.ParseUint(c.config.ID, 10, 32)
		if err != nil {
			log.Errorf("invalid client id: %v", err)
			time.Sleep(c.config.RetriesDelay)
			continue
		}

		c.createClientSocket()
		if c.conn == nil {
			log.Errorf("action: connect | result: fail | client_id: %v", c.config.ID)
			time.Sleep(c.config.RetriesDelay)
			continue
		}
		log.Infof("action: connect | result: success | client_id: %v", c.config.ID)

		encodedMsg, err := encodeResultsMessage(uint32(id))
		if err != nil {
			log.Errorf("action: encode_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
			continue
		}

		if err := sendAll(c.conn, encodedMsg); err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
			c.conn.Close()
			time.Sleep(c.config.RetriesDelay)
			continue
		}
		log.Infof("action: send_message | result: success | client_id: %v", c.config.ID)

		status_code := make([]byte, 1)
		if _, err := io.ReadFull(c.conn, status_code); err != nil {
			log.Errorf("action: receive_message | result: fail | error: %v", err)
			c.conn.Close()
			time.Sleep(c.config.RetriesDelay)
			continue
		}

		if status_code[0] == ResponseError {
			log.Errorf("action: received_message | result: fail | client_id: %v | response: %v", c.config.ID, status_code[0])
			c.conn.Close()
			time.Sleep(c.config.RetriesDelay)
			continue
		}

		winners, err := decodeResultsResponse(c.conn)
		c.conn.Close()
		if err != nil {
			log.Errorf("action: received_message 2 | result: fail | client_id: %v | error: %v", c.config.ID, err)
			time.Sleep(c.config.RetriesDelay)
			continue
		}

		for _, w := range winners {
			log.Infof("action: ganador | result: success | client_id: %v | ganador_dni: %v", id, w)
		}

		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v",
			len(winners),
		)
		return
	}
}


