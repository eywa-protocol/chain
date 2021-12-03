package validation

import (
	"errors"
	"fmt"

	"github.com/eywa-protocol/chain/core/types"
)

// VerifyBlock checks whether the block is valid
// func VerifyBlock(block *types.Block, ld *ledger.Ledger, completely bool) error {
// 	header := block.Header
// 	if header.Height == 0 {
// 		return nil
// 	}

// 	m := len(header.EpochValidators) - (len(header.EpochValidators)-1)/3
// 	hash := block.Hash()
// 	err := signature.VerifyMultiSignature(hash[:], header.EpochValidators, m, header.SigData)
// 	if err != nil {
// 		return err
// 	}

// 	prevHeader, err := ld.GetHeaderByHash(block.Header.PrevBlockHash)
// 	if err != nil {
// 		return fmt.Errorf("[BlockValidator], can not find prevHeader: %s", err)
// 	}

// 	err = VerifyHeader(block.Header, prevHeader)
// 	if err != nil {
// 		return err
// 	}

// 	//verfiy block's transactions
// 	if completely {

// 		//TODO: NextEpoch Check.
// 		/*		bookkeeperaddress, err := ledger.GetEpochAddress(ld.Blockchain.GetEpochValidatorsByTXs(block.Transactions))
// 				if err != nil {
// 					return errors.New(fmt.Sprintf("GetEpochAddress Failed."))
// 				}
// 				if block.Header.NextEpoch != bookkeeperaddress {
// 					return errors.New(fmt.Sprintf("Epoch is not validate."))
// 				}*/

// 		for _, txVerify := range block.Transactions {
// 			if errCode := VerifyTransaction(txVerify); errCode != ontErrors.ErrNoError {
// 				return errors.New(fmt.Sprintf("VerifyTransaction failed when verifiy block"))
// 			}

// 			if errCode := VerifyTransactionWithLedger(txVerify, ld); errCode != ontErrors.ErrNoError {
// 				return errors.New(fmt.Sprintf("VerifyTransaction failed when verifiy block"))
// 			}
// 		}
// 	}

// 	return nil
// }

func VerifyHeader(header, prevHeader *types.Header) error {
	if header.Height == 0 {
		return nil
	}

	if prevHeader == nil {
		return errors.New("[BlockValidator], can not find previous block.")
	}

	if prevHeader.Height+1 != header.Height {
		return errors.New("[BlockValidator], block height is incorrect.")
	}

	if prevHeader.Timestamp >= header.Timestamp {
		return errors.New("[BlockValidator], block timestamp is incorrect.")
	}

	address, err := types.AddressFromPubLeySlice(header.EpochValidators)
	if err != nil {
		return err
	}

	if prevHeader.NextEpoch != address {
		return fmt.Errorf("bookkeeper address error")
	}

	return nil
}
