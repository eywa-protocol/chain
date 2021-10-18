/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */

package ledgerstore

import (
	"fmt"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/payload"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/store"
	scommon "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/store/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/store/overlaydb"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native/event"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native/storage"
)

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

func SaveNotify(eventStore scommon.EventStore, txHash common.Uint256, notify *event.ExecuteNotify) error {
	if !config.DefConfig.Common.EnableEventLog {
		return nil
	}
	if err := eventStore.SaveEventNotifyByTx(txHash, notify); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	// TODO add event
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notify)
	return nil
}
