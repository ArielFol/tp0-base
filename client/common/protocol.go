package common

import (
	"bytes"
	"encoding/binary"
	"io"
)

const (
	ResponseOk    uint8 = 1
	ResponseError   uint8 = 2
)

const (
	MessageTypeBets    uint8 = 1
	MessageTypeFinish  uint8 = 2
	MessageTypeResults   uint8 = 3
)

//EncodeBet Serializes a Bet struct into a byte slice using big-endian encoding:
// - message type: 1 byte (uint8) with value 1 for bets messages
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

	if err := binary.Write(buffer, binary.BigEndian, uint8(MessageTypeBets)); err != nil {
		return nil, err
	}

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

func encodeFinishMessage(agency uint32) ([]byte, error) {
	buffer := new(bytes.Buffer)

	if err := binary.Write(buffer, binary.BigEndian, uint8(MessageTypeFinish)); err != nil {
		return nil, err
	}

	if err := binary.Write(buffer, binary.BigEndian, agency); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func encodeResultsMessage(agency uint32) ([]byte, error) {
	buffer := new(bytes.Buffer)
	
	if err := binary.Write(buffer, binary.BigEndian, uint8(MessageTypeResults)); err != nil {
		return nil, err
	}

	if err := binary.Write(buffer, binary.BigEndian, agency); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeResultsResponse(reader io.Reader) ([]uint64, error) {
	var winnersCount uint32
	if err := binary.Read(reader, binary.BigEndian, &winnersCount); err != nil {
		return nil, err
	}

	winners := make([]uint64, winnersCount)
	for i := uint32(0); i < winnersCount; i++ {
		var winnerDNI uint64
		if err := binary.Read(reader, binary.BigEndian, &winnerDNI); err != nil {
			return nil, err
		}
		winners[i] = winnerDNI
	}

	return winners, nil
}


