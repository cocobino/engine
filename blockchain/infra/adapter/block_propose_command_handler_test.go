/*
 * Copyright 2018 It-chain
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package adapter_test

import (
	"testing"

	"os"
	"sync"

	"github.com/it-chain/engine/blockchain"
	"github.com/it-chain/engine/blockchain/infra/adapter"
	"github.com/it-chain/engine/blockchain/infra/mem"
	"github.com/it-chain/engine/common/command"

	"time"

	"github.com/it-chain/engine/blockchain/api"
	"github.com/it-chain/engine/blockchain/infra/repo"
	"github.com/it-chain/engine/blockchain/test/mock"
	"github.com/it-chain/engine/common"
	"github.com/it-chain/engine/common/event"
	"github.com/it-chain/engine/common/rabbitmq/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestBlockProposeCommandHandler_HandleProposeBlockCommand_Solo(t *testing.T) {

	//set subscriber
	var wg sync.WaitGroup
	wg.Add(2)

	subscriber := pubsub.NewTopicSubscriber("", "Event")
	defer subscriber.Close()

	handler := &mock.CommitEventHandler{}
	handler.HandleFunc = func(event event.BlockCommitted) {
		assert.Equal(t, "tx01", event.TxList[0].ID)
		assert.Equal(t, blockchain.Committed, event.State)
		wg.Done()
	}

	subscriber.SubscribeTopic("block.*", handler)

	//set bApi
	publisherID := "junksound"
	dbPath := "./.db"

	br, err := repo.NewBlockRepository(dbPath)
	assert.Equal(t, nil, err)
	defer func() {
		br.Close()
		os.RemoveAll(dbPath)
	}()

	prevBlock := mock.GetNewBlock([]byte("genesis"), 0)

	err = br.AddBlock(prevBlock)
	assert.NoError(t, err)

	eventService := common.NewEventService("", "Event")
	blockPool := mem.NewBlockPool()

	bApi, err := api.NewBlockApi(publisherID, br, eventService, blockPool)
	assert.NoError(t, err)

	commandHandler := adapter.NewBlockProposeCommandHandler(bApi, "solo")

	//when
	err = commandHandler.HandleProposeBlockCommand(command.ProposeBlock{TxList: nil})
	//then
	assert.Error(t, err)

	//when
	err = commandHandler.HandleProposeBlockCommand(command.ProposeBlock{TxList: make([]command.Tx, 0)})
	//then
	assert.Error(t, err)

	//when
	err = commandHandler.HandleProposeBlockCommand(command.ProposeBlock{
		TxList: []command.Tx{
			{
				ID:        "tx01",
				ICodeID:   "ICodeID",
				PeerID:    "2",
				TimeStamp: time.Now().Round(0),
				Jsonrpc:   "123",
				Function:  "function1",
				Args:      []string{"arg1", "arg2"},
				Signature: []byte{0x1},
			},
		},
	})

	//then
	assert.NoError(t, err)

	//when
	err = commandHandler.HandleProposeBlockCommand(command.ProposeBlock{
		TxList: []command.Tx{
			{
				ID:        "tx01",
				ICodeID:   "ICodeID",
				PeerID:    "2",
				TimeStamp: time.Now().Round(0),
				Jsonrpc:   "123",
				Function:  "function1",
				Args:      []string{"arg1", "arg2"},
				Signature: []byte{0x1},
			},

			{
				ID:        "tx02",
				ICodeID:   "ICodeID",
				PeerID:    "2",
				TimeStamp: time.Now().Round(0),
				Jsonrpc:   "123",
				Function:  "function1",
				Args:      []string{"arg1", "arg2"},
				Signature: []byte{0x1},
			},
		},
	})

	//then
	assert.NoError(t, err)

	wg.Wait()
}

func TestBlockProposeCommandHandler_HandleProposeBlockCommand_PBFT_OneTransaction(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	subscriber := pubsub.NewTopicSubscriber("", "Event")
	defer subscriber.Close()

	handler := &mock.StartConsensusCommandHandler{}
	handler.HandleFunc = func(command command.StartConsensus) {
		assert.Equal(t, "tx01", command.TxList[0].ID)
		assert.Equal(t, blockchain.Created, command.State)
		wg.Done()
	}

	subscriber.SubscribeTopic("block.*", handler)

	dbPath := "./.db"
	blockRepository, err := repo.NewBlockRepository(dbPath)
	assert.Equal(t, nil, err)
	defer func() {
		blockRepository.Close()
		os.RemoveAll(dbPath)
	}()

	prevBlock := mock.GetNewBlock([]byte("genesis"), 0)
	err = blockRepository.AddBlock(prevBlock)
	assert.NoError(t, err)

	publisherID := "iAmPublisher"
	eventService := common.NewEventService("", "Event")
	blockPool := mem.NewBlockPool()

	api, err := api.NewBlockApi(publisherID, blockRepository, eventService, blockPool)
	assert.NoError(t, err)

	commandHandler := adapter.NewBlockProposeCommandHandler(api, "pbft")

	//when
	err = commandHandler.HandleProposeBlockCommand(command.ProposeBlock{
		TxList: []command.Tx{
			{
				ID:        "tx01",
				ICodeID:   "ICodeID",
				PeerID:    "2",
				TimeStamp: time.Now().Round(0),
				Jsonrpc:   "123",
				Function:  "function1",
				Args:      []string{"arg1", "arg2"},
				Signature: []byte{0x1},
			},
		},
	})
	//then
	assert.NoError(t, err)

	wg.Wait()
}

func TestBlockProposeCommandHandler_HandleProposeBlockCommand_PBFT_TwoTransaction(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	subscriber := pubsub.NewTopicSubscriber("", "Event")
	defer subscriber.Close()

	handler := &mock.StartConsensusCommandHandler{}
	handler.HandleFunc = func(command command.StartConsensus) {
		assert.Equal(t, "tx01", command.TxList[0].ID)
		assert.Equal(t, blockchain.Created, command.State)
		wg.Done()
	}

	subscriber.SubscribeTopic("block.*", handler)

	dbPath := "./.db"
	blockRepository, err := repo.NewBlockRepository(dbPath)
	assert.Equal(t, nil, err)
	defer func() {
		blockRepository.Close()
		os.RemoveAll(dbPath)
	}()

	prevBlock := mock.GetNewBlock([]byte("genesis"), 0)
	err = blockRepository.AddBlock(prevBlock)
	assert.NoError(t, err)

	publisherID := "iAmPublisher"
	eventService := common.NewEventService("", "Event")
	blockPool := mem.NewBlockPool()

	api, err := api.NewBlockApi(publisherID, blockRepository, eventService, blockPool)
	assert.NoError(t, err)

	commandHandler := adapter.NewBlockProposeCommandHandler(api, "pbft")

	//when
	err = commandHandler.HandleProposeBlockCommand(command.ProposeBlock{
		TxList: []command.Tx{
			{
				ID:        "tx01",
				ICodeID:   "ICodeID",
				PeerID:    "2",
				TimeStamp: time.Now().Round(0),
				Jsonrpc:   "123",
				Function:  "function1",
				Args:      []string{"arg1", "arg2"},
				Signature: []byte{0x1},
			},
			{
				ID:        "tx02",
				ICodeID:   "ICodeID",
				PeerID:    "2",
				TimeStamp: time.Now().Round(0),
				Jsonrpc:   "123",
				Function:  "function1",
				Args:      []string{"arg1", "arg2"},
				Signature: []byte{0x1},
			},
		},
	})
	//then
	assert.NoError(t, err)

	wg.Wait()
}

func TestBlockProposeCommandHandler_HandleProposeBlockCommand_PBFT_ExceptionCases(t *testing.T) {
	//ErrConsesnsusService := errors.New("Consensus Error")

	dbPath := "./.db"
	blockRepository, err := repo.NewBlockRepository(dbPath)
	assert.Equal(t, nil, err)
	defer func() {
		blockRepository.Close()
		os.RemoveAll(dbPath)
	}()

	prevBlock := mock.GetNewBlock([]byte("genesis"), 0)
	err = blockRepository.AddBlock(prevBlock)
	assert.NoError(t, err)

	publisherID := "iAmPublisher"
	eventService := common.NewEventService("", "Event")
	blockPool := mem.NewBlockPool()

	api, err := api.NewBlockApi(publisherID, blockRepository, eventService, blockPool)
	assert.NoError(t, err)

	commandHandler := adapter.NewBlockProposeCommandHandler(api, "pbft")

	//when
	err = commandHandler.HandleProposeBlockCommand(command.ProposeBlock{TxList: nil})
	//then
	assert.Error(t, err)

	//when
	err = commandHandler.HandleProposeBlockCommand(command.ProposeBlock{TxList: make([]command.Tx, 0)})
	//then
	assert.Error(t, err)

	//when
	//when
	err = commandHandler.HandleProposeBlockCommand(command.ProposeBlock{
		TxList: []command.Tx{
			{
				ID:        "tx01",
				ICodeID:   "ICodeID",
				PeerID:    "2",
				TimeStamp: time.Now().Round(0),
				Jsonrpc:   "123",
				Function:  "function1",
				Args:      []string{"arg1", "arg2"},
				Signature: []byte{0x1},
			},
		},
	})
	//then
	assert.NoError(t, err)

}
