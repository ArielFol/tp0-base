package common

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"encoding/csv"
	"io"
)

const MaxBatchBytes = 8 * 1024

type Bet struct {
	Agency 		uint32
	Name  		string
	Surname 	string
	DNI			uint64
	Birthdate 	int64
	Number 		uint32
}

type BetReader struct {
	reader *csv.Reader
	pending Bet
	hasPending bool
}

func NewBet(agency uint32) (*Bet, error) {
	name := os.Getenv("NOMBRE")
	surname := os.Getenv("APELLIDO")
	dniStr := os.Getenv("DOCUMENTO")
	birthdateStr := os.Getenv("NACIMIENTO")
	numberStr := os.Getenv("NUMERO")

	if name == "" || surname == "" || dniStr == "" || birthdateStr == "" || numberStr == "" {
		return nil, fmt.Errorf("missing required env vars")
	}

	dni, err := strconv.ParseUint(dniStr, 10, 64)
	if err != nil {
		return nil, err
	}

	number, err := strconv.ParseUint(numberStr, 10, 32)
	if err != nil {
		return nil, err
	}

	birthdate, err := time.Parse("2006-01-02", birthdateStr)
	if err != nil {
		return nil, err
	}

	return &Bet{
		Agency: agency,
		Name: name,
		Surname: surname,
		DNI: dni,
		Birthdate: birthdate.Unix(),
		Number: uint32(number),
	}, nil
}

func readNextBets(betReader *BetReader, max int, agency uint32) ([]Bet, error) {
	var bets []Bet
	currentSize := 0

	for i := 0; i < max; i++ {
		log.Infof("reading bet %v for agency %v", i+1, agency)

		var bet Bet

		if betReader.hasPending {
			bet = betReader.pending
			betReader.hasPending = false
		} else {
			log.Infof("there is no pending bet for agency %v, reading from file", agency)
			record, err := betReader.reader.Read()
			if err != nil {
				if err == io.EOF {
					if len(bets) > 0 {
						return bets, nil
					}
				}
				return nil, err
			}

			dni, err := strconv.ParseUint(record[2], 10, 64)
			if err != nil {
				return nil, err
			}

			birthdate, err := time.Parse("2006-01-02", record[3])
			if err != nil {
				return nil, err
			}

			number, err := strconv.ParseUint(record[4], 10, 32)
			if err != nil {
				return nil, err
			}

			bet = Bet{
				Agency:		agency,
				Name:      	record[0],
				Surname:   	record[1],
				DNI:       	dni,
				Birthdate: 	birthdate.Unix(),
				Number:    	uint32(number),
			}

			encodedBet, err := EncodeBet(&bet)
			if err != nil {
				return nil, err
			}

			if len(bets) == 0 {
				// header
				currentSize += 4
			}
			
			if currentSize + len(encodedBet) > MaxBatchBytes {
				betReader.pending = bet
				betReader.hasPending = true
				break
			}

			bets = append(bets, bet)
			currentSize += len(encodedBet)
		}
	}
	if len(bets) == 0 {
		log.Infof("action: read_bets | result: EOF | agency: %v", agency)
		return nil, io.EOF
	}

	return bets, nil
}

