package util

import (
	"encoding/hex"
)

func StringToHex(hexHash string) ([]byte, error) {
	// Step 1: Decode the hex string into a byte slice
	hashBytes, err := hex.DecodeString(hexHash)
	if err != nil {
		// Log the error and return a nil slice and the error
		Logger.Error().Msgf("failed to decode hex string: %s", err)
		return nil, err
	}

	// Return the decoded []byte
	return hashBytes, nil
}

func StringToHexInt(hexHash string) ([]int, error) {

	// Step 1: Decode the hex string into a byte slice
	hashBytes, err := hex.DecodeString(hexHash)
	if err != nil {
		Logger.Error().Msgf("failed to decode hex string: %s", err)
		return nil, err
	}

	var decimalRepresentation []int
	for _, byteValue := range hashBytes {
		decimalRepresentation = append(decimalRepresentation, int(byteValue))
	}

	return decimalRepresentation, nil
}
