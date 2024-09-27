package loader

import (
	"bufio"
	"encoding/hex"
	"os"
	"strconv"
	"strings"

	"github.com/unisat-wallet/libbrc20-indexer-fractal/model"
)

func LoadBRC20InputData(fname string) []*model.InscriptionBRC20Data {
	brc20Data := make([]*model.InscriptionBRC20Data, 0, 10240)

	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	max := 128 * 1024 * 1024
	buf := make([]byte, max)
	scanner.Buffer(buf, max)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, " ")

		if len(fields) != 14 {
			panic("invalid data format")
		}

		var data model.InscriptionBRC20Data
		data.IsTransfer, err = strconv.ParseBool(fields[0])
		if err != nil {
			panic(err)
		}

		txid, err := hex.DecodeString(fields[1])
		if err != nil {
			panic(err)
		}
		data.TxId = string(txid)

		idx, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			panic(err)
		}
		data.Idx = uint32(idx)

		vout, err := strconv.ParseUint(fields[3], 10, 32)
		if err != nil {
			panic(err)
		}
		data.Vout = uint32(vout)

		offset, err := strconv.ParseUint(fields[4], 10, 64)
		if err != nil {
			panic(err)
		}
		data.Offset = uint64(offset)

		satoshi, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			panic(err)
		}
		data.Satoshi = uint64(satoshi)

		pkScript, err := hex.DecodeString(fields[6])
		if err != nil {
			panic(err)
		}
		data.PkScript = string(pkScript)

		inscriptionNumber, err := strconv.ParseInt(fields[7], 10, 64)
		if err != nil {
			panic(err)
		}
		data.InscriptionNumber = int64(inscriptionNumber)

		data.ContentBody, err = hex.DecodeString(fields[8])
		if err != nil {
			panic(err)
		}

		createIdxKey, err := strconv.ParseUint(fields[9], 16, 64)
		if err != nil {
			panic(err)
		}
		data.CreateIdxKey = uint64(createIdxKey)

		createIdxKeyStr, err := hex.DecodeString(fields[9])
		if err != nil {
			panic(err)
		}
		data.CreateIdxString = string(createIdxKeyStr)

		height, err := strconv.ParseUint(fields[10], 10, 32)
		if err != nil {
			panic(err)
		}
		data.Height = uint32(height)

		txIdx, err := strconv.ParseUint(fields[11], 10, 32)
		if err != nil {
			panic(err)
		}
		data.TxIdx = uint32(txIdx)

		blockTime, err := strconv.ParseUint(fields[12], 10, 32)
		if err != nil {
			panic(err)
		}
		data.BlockTime = uint32(blockTime)

		sequence, err := strconv.ParseUint(fields[13], 10, 16)
		if err != nil {
			panic(err)
		}
		data.Sequence = uint16(sequence)

		brc20Data = append(brc20Data, &data)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return brc20Data
}
