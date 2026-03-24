package common

import (
	"bytes"
	"encoding/binary"
)

//EncodeBet Serializes a Bet struct into a byte slice using big-endian encoding:
// - Agency: 4 bytes (uint32)
// - Name Length: 4 bytes (uint32) followed by Name bytes
// - Surname Length: 4 bytes (uint32) followed by Surname bytes
// - DNI: 8 bytes (uint64)
// - Birthdate: 8 bytes (int64)
// - Number: 4 bytes (uint32)

func EncodeBet(bet *Bet) ([]byte, error) {
	buffer:= new(bytes.Buffer)

	if err := binary.Write(buffer, binary.BigEndian, bet.Agency); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, uint32(len(bet.Name))); err != nil {
		return nil, err
	}
	if _, err := buffer.Write([]byte(bet.Name)); err != nil {
		return nil, err
	}

	if err := binary.Write(buffer, binary.BigEndian, uint32(len(bet.Surname))); err != nil {
		return nil, err
	}

	if _, err := buffer.Write([]byte(bet.Surname)); err != nil {
		return nil, err
	}

	if err := binary.Write(buffer, binary.BigEndian, bet.DNI); err != nil {
		return nil, err
	}
	
	if err := binary.Write(buffer, binary.BigEndian, bet.Birthdate); err != nil {
		return nil, err
	}
	if err := binary.Write(buffer, binary.BigEndian, bet.Number); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func encodeBets(bets []Bet) ([]byte, error) {
	buffer := new(bytes.Buffer)

	if err := binary.Write(buffer, binary.BigEndian, uint32(len(bets))); err != nil {
		return nil, err
	}
	
	for _, bet := range bets {
		encodedBet, err := EncodeBet(&bet)
		if err != nil {
			return nil, err
		}

		if _, err := buffer.Write(encodedBet); err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}