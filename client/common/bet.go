package common

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Bet struct {
	Agency 		uint32
	Name  		string
	Surname 	string
	DNI			uint64
	Birthdate 	int64
	Number 		uint32
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

