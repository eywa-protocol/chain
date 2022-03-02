package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func GetJsonObjectFromFile(filePath string, jsonObject interface{}) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	// Remove the UTF-8 Byte Order Mark
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	err = json.Unmarshal(data, jsonObject)
	if err != nil {
		return fmt.Errorf("json.Unmarshal %s error:%s", data, err)
	}
	return nil
}

func GetStoreDirPath(dataDir, networkName string) string {
	return dataDir + string(os.PathSeparator) + networkName
}

func GenExportBlocksFileName(name string, start, end uint32) string {
	index := strings.LastIndex(name, ".")
	fileName := ""
	fileExt := ""
	if index < 0 {
		fileName = name
	} else {
		fileName = name[0:index]
		if index < len(name)-1 {
			fileExt = name[index+1:]
		}
	}
	fileName = fmt.Sprintf("%s_%d_%d", fileName, start, end)
	if index > 0 {
		fileName = fileName + "." + fileExt
	}
	return fileName
}

/*func SendRawTransactionData(txData string) (string, error) {
	data, ontErr := sendRpcRequest("sendrawtransaction", []interface{}{txData})
	if ontErr != nil {
		return "", ontErr.Error
	}
	hexHash := ""
	err := json.Unmarshal(data, &hexHash)
	if err != nil {
		return "", fmt.Errorf("json.Unmarshal hash:%s error:%s", data, err)
	}
	return hexHash, nil
}

func PrepareSendRawTransaction(txData string) (*states.PreExecResult, error) {
	data, ontErr := sendRpcRequest("sendrawtransaction", []interface{}{txData, 1})
	if ontErr != nil {
		return nil, ontErr.Error
	}
	preResult := &states.PreExecResult{}
	err := json.Unmarshal(data, &preResult)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal PreExecResult:%s error:%s", data, err)
	}
	return preResult, nil
}

func GetRawTransaction(txHash string) ([]byte, error) {
	data, ontErr := sendRpcRequest("getrawtransaction", []interface{}{txHash, 1})
	if ontErr == nil {
		return data, nil
	}
	switch ontErr.ErrorCode {
	case ERROR_INVALID_PARAMS:
		return nil, fmt.Errorf("invalid TxHash:%s", txHash)
	}
	return nil, ontErr.Error
}

func GetBlock(hashOrHeight interface{}) ([]byte, error) {
	data, ontErr := sendRpcRequest("getblock", []interface{}{hashOrHeight, 1})
	if ontErr == nil {
		return data, nil
	}
	switch ontErr.ErrorCode {
	case ERROR_INVALID_PARAMS:
		return nil, fmt.Errorf("invalid block hash or block height:%v", hashOrHeight)
	}
	return nil, ontErr.Error
}

func GetNetworkId() (uint32, error) {
	data, ontErr := sendRpcRequest("getnetworkid", []interface{}{})
	if ontErr != nil {
		return 0, ontErr.Error
	}
	var networkId uint32
	err := json.Unmarshal(data, &networkId)
	if err != nil {
		return 0, fmt.Errorf("json.Unmarshal networkId error:%s", err)
	}
	return networkId, nil
}

func GetBlockData(hashOrHeight interface{}) ([]byte, error) {
	data, ontErr := sendRpcRequest("getblock", []interface{}{hashOrHeight})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, fmt.Errorf("invalid block hash or block height:%v", hashOrHeight)
		}
		return nil, ontErr.Error
	}
	hexStr := ""
	err := json.Unmarshal(data, &hexStr)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal error:%s", err)
	}
	blockData, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	return blockData, nil
}

func GetBlockCount() (uint32, error) {
	data, ontErr := sendRpcRequest("getblockcount", []interface{}{})
	if ontErr != nil {
		return 0, ontErr.Error
	}
	num := uint32(0)
	err := json.Unmarshal(data, &num)
	if err != nil {
		return 0, fmt.Errorf("json.Unmarshal:%s error:%s", data, err)
	}
	return num, nil
}

func GetTxHeight(txHash string) (uint32, error) {
	data, ontErr := sendRpcRequest("getblockheightbytxhash", []interface{}{txHash})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return 0, fmt.Errorf("cannot find tx by:%s", txHash)
		}
		return 0, ontErr.Error
	}
	height := uint32(0)
	err := json.Unmarshal(data, &height)
	if err != nil {
		return 0, fmt.Errorf("json.Unmarshal error:%s", err)
	}
	return height, nil
}

func SignTransaction(signer *account.Account, tx *types.Transaction) error {
	txHash := tx.Hash()
	tx.Sig.SigData = signer.PrivateKey.Sign(txHash.ToArray())
	sigData, err := Sign(txHash.ToArray(), signer)
	if err != nil {
		return fmt.Errorf("sign error:%s", err)
	}
	hasSig := false
	for i, sig := range tx.Sigs {
		if len(sig.PubKeys) == 1 && pubKeysEqual(sig.PubKeys, []bls.PublicKey{signer.PublicKey}) {
			if hasAlreadySig(txHash.ToArray(), signer.PublicKey, sig.SigData) {
				//has already signed
				return nil
			}
			hasSig = true
			//replace
			tx.Sigs[i].SigData = [][]byte{sigData}
		}
	}
	if !hasSig {
		tx.Sigs = append(tx.Sigs, types.Sig{
			PubKeys: []bls.PublicKey{signer.PublicKey},
			M:       1,
			SigData: [][]byte{sigData},
		})
	}
	return nil
}

func MultiSigTransaction(mutTx *types.Transaction, mk bls.Signature, allPub bls.PublicKey, signer *account.Account) error {
	pkSize := len(pubKeys)
	if m == 0 || int(m) > pkSize || pkSize > constants.MULTI_SIG_MAX_PUBKEY_SIZE {
		return fmt.Errorf("invalid params")
	}

	validPubKey := false
	for _, pk := range pubKeys {
		if keypair.ComparePublicKey(pk, signer.PublicKey.Marshal()) {
			validPubKey = true
			break
		}
	}
	if !validPubKey {
		return fmt.Errorf("invalid signer")
	}

	if len(mutTx.Sigs) == 0 {
		mutTx.Sigs = make([]types.Sig, 0)
	}

	m := uint64(1) << signer.Id
	if mutTx.Sig.M&m != 0 {
		return errors.New("MultiSigTransaction attempt to sign one more time")
	}
	mutTx.Sig.M |= m
	txHash := mutTx.Hash()
	sig := signer.PrivateKey.Multisign(txHash.ToArray(), allPub, mk)
	mutTx.Sig.PubKey = mutTx.Sig.PubKey.Aggregate(signer.PublicKey)
	mutTx.Sig.SigData = mutTx.Sig.SigData.Aggregate(sig)

	sigData, err := Sign(txHash.ToArray(), signer)
	if err != nil {
		return fmt.Errorf("sign error:%s", err)
	}

	hasMutilSig := false
	for i, sigs := range mutTx.Sigs {
		if !pubKeysEqual(sigs.PubKeys, pubKeys) {
			continue
		}
		hasMutilSig = true
		if hasAlreadySig(txHash.ToArray(), signer.PublicKey, sigs.SigData) {
			break
		}
		sigs.SigData = append(sigs.SigData, sigData)
		mutTx.Sigs[i] = sigs
		break
	}
	if !hasMutilSig {
		mutTx.Sigs = append(mutTx.Sigs, types.Sig{
			PubKeys: pubKeys,
			M:       uint16(m),
			SigData: [][]byte{sigData},
		})
	}
	return nil
}*/

/*func GetSmartContractEventInfo(txHash string) ([]byte, error) {
	data, ontErr := sendRpcRequest("getsmartcodeevent", []interface{}{txHash})
	if ontErr == nil {
		return data, nil
	}
	switch ontErr.ErrorCode {
	case ERROR_INVALID_PARAMS:
		return nil, fmt.Errorf("invalid TxHash:%s", txHash)
	}
	return nil, ontErr.Error
}*/

/*func hasAlreadySig(data []byte, pk bls.PublicKey, sigDatas [][]byte) bool {
	for _, sigData := range sigDatas {
		err := signature.Verify(pk, data, sigData)
		if err == nil {
			return true
		}
	}
	return false
}*/

/*func pubKeysEqual(pks1, pks2 []bls.PublicKey) bool {
	if len(pks1) != len(pks2) {
		return false
	}
	size := len(pks1)
	if size == 0 {
		return true
	}
	pkstr1 := make([]string, 0, size)
	for _, pk := range pks1 {
		pkstr1 = append(pkstr1, hex.EncodeToString(pk.Marshal()))
	}
	pkstr2 := make([]string, 0, size)
	for _, pk := range pks2 {
		pkstr2 = append(pkstr2, hex.EncodeToString(pk.Marshal()))
	}
	sort.Strings(pkstr1)
	sort.Strings(pkstr2)
	for i := 0; i < size; i++ {
		if pkstr1[i] != pkstr2[i] {
			return false
		}
	}
	return true
}*/

/*func Sign(data []byte, signer *account.Account) (bls.Signature, error) {

	s := signer.PrivateKey.Sign(data)

	//sigData := s.Marshal()

	return s, nil
}*/
