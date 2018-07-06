package txpool

import (
	"errors"
	"log"
)

type TxpoolQueryService interface {
	GetLeader() Leader
	GetAllStagedTransactions() ([]Transaction, error)
	GetAllCreatedTransactions() ([]Transaction, error)
}

type TransferService interface {
	SendTransactionsToLeader(transactions []Transaction, leader Leader) error
}

type BlockService interface {
	ProposeBlock(transactions []Transaction) error
}

type TxTransferService struct {
	txpoolQueryService TxpoolQueryService
	transferService    TransferService
}

func NewTxPeriodicTransferService(queryService TxpoolQueryService, transferService TransferService) *TxTransferService {

	return &TxTransferService{
		txpoolQueryService: queryService,
		transferService:    transferService,
	}
}

func (t TxTransferService) TransferCreatedTxToLeader() error {

	transactions, err := t.txpoolQueryService.GetAllCreatedTransactions()

	if err != nil {
		log.Println(err.Error())
		return err
	}

	leader := t.txpoolQueryService.GetLeader()

	if leader.LeaderId.ToString() == "" {
		return errors.New("there is no leader")
	}

	err = t.transferService.SendTransactionsToLeader(transactions, leader)

	//if err != nil {
	//	log.Println(err.Error())
	//	return err
	//}

	//if err := t.removeTxs(transactions); err != nil {
	//	log.Println(err.Error())
	//	return err
	//}

	return nil
}

//func (t TxPeriodicTransferService) removeTxs(transactions []*Transaction) error {
//
//	for _, tx := range transactions {
//		if err := t.txRepository.Remove(tx.TxId); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
