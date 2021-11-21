/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
 */package ledgerstore

import (
	"fmt"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/payload"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/store"
	scommon "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/store/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/store/overlaydb"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native/event"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native/storage"
)

//HandleAnyTransaction deal with smart contract
func (self *StateStore) HandleBridgeTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB, cache *storage.CacheDB,
	tx *types.Transaction, block *types.Block, notify *event.ExecuteNotify) ([]common.Uint256, error) {
	be := tx.Payload.(*payload.BridgeEvent)
	beBytes := common.SerializeToBytes(be)
	service, err := native.NewNativeService(cache, tx, block.Header.Timestamp, block.Header.Height,
		block.Hash(), block.Header.ChainID, beBytes, false)
	if err != nil {
		return nil, fmt.Errorf("HandleInvokeTransaction Error: %+v\n", err)
	}
	if _, err := service.Invoke(); err != nil {
		return nil, err
	}
	notify.Notify = append(notify.Notify, service.GetNotify()...)
	notify.State = event.CONTRACT_STATE_SUCCESS
	service.GetCacheDB().Commit()
	return service.GetCrossHashes(), nil
}

//HandleInvokeTransaction deal with smart contract invoke transaction
func (self *StateStore) HandleInvokeTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB, cache *storage.CacheDB,
	tx *types.Transaction, block *types.Block, notify *event.ExecuteNotify) ([]common.Uint256, error) {
	invoke := tx.Payload.(*payload.InvokeCode)
	service, err := native.NewNativeService(cache, tx, block.Header.Timestamp, block.Header.Height,
		block.Hash(), block.Header.ChainID, invoke.Code, false)
	if err != nil {
		return nil, fmt.Errorf("HandleInvokeTransaction Error: %+v\n", err)
	}
	if _, err := service.Invoke(); err != nil {
		return nil, err
	}
	notify.Notify = append(notify.Notify, service.GetNotify()...)
	notify.State = event.CONTRACT_STATE_SUCCESS
	service.GetCacheDB().Commit()
	return service.GetCrossHashes(), nil
}

func (self *StateStore) HandleEpochTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB, cache *storage.CacheDB,
	tx *types.Transaction, block *types.Block, notify *event.ExecuteNotify) ([]common.Uint256, error) {
	var epoch = tx.Payload.(*payload.Epoch)
	service, err := native.NewNativeService(cache, tx, block.Header.Timestamp, block.Header.Height,
		block.Hash(), block.Header.ChainID, epoch.Data, false)
	if err != nil {
		return nil, fmt.Errorf("HandleInvokeTransaction Error: %+v\n", err)
	}
	notify.Notify = append(notify.Notify, service.GetNotify()...)
	notify.State = event.CONTRACT_STATE_SUCCESS
	service.GetCacheDB().Commit()
	return service.GetCrossHashes(), nil
}

func SaveNotify(eventStore scommon.EventStore, txHash common.Uint256, notify *event.ExecuteNotify) error {
	if err := eventStore.SaveEventNotifyByTx(txHash, notify); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	// TODO add event
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notify)
	return nil
}
