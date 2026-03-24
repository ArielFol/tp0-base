package common

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"encoding/csv"
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
	pending *Bet
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

func readNextBets(betReader *BetReader, max int) ([]*Bet, error) {
	var bets []*Bet
	currentSize := 0

	for i := 0; i < max; i++ {
		if betReader.pending != nil {
			bet = betReader.pending
			betReader.pending = nil
		} else {
			

			record, err := betReader.reader.Read()
			if err != nil {
				return nil, err
			}

			agency, err := strconv.ParseUint(record[0], 10, 32)
			if err != nil {
				return nil, err
			}

			dni, err := strconv.ParseUint(record[3], 10, 64)
			if err != nil {
				return nil, err
			}

			birthdate, err := strconv.ParseInt(record[4], 10, 64)
			if err != nil {
				return nil, err
			}

			number, err := strconv.ParseUint(record[5], 10, 32)
			if err != nil {
				return nil, err
			}

			bet := Bet{
				Agency:    uint32(agency),
				Name:      record[1],
				Surname:   record[2],
				DNI:       dni,
				Birthdate: birthdate,
				Number:    uint32(number),
			}
		}

		encodedBet, err := EncodeBet(&bet)
		if err != nil {
			return nil, err
		}

		if len(bets) == 0 {
			currentSize += 4
		}
		
		if currentSize + len(encodedBet) > MaxBatchBytes {
			betReader.pending = &bet
			break
		}

		bets = append(bets, bet)
	}

	if len(bets) == 0 {
		return nil, io.EOF
	}

	return bets, nil
}

