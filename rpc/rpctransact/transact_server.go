package rpctransact

import (
	"fmt"

	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"golang.org/x/net/context"
)

type transactServer struct {
	transactor *execution.Transactor
	txCodec    txs.Codec
}

func NewTransactServer(transactor *execution.Transactor, txCodec txs.Codec) TransactServer {
	return &transactServer{
		transactor: transactor,
		txCodec:    txCodec,
	}
}

func (ts *transactServer) BroadcastTxSync(ctx context.Context, param *TxEnvelopeParam) (*exec.TxExecution, error) {
	txEnv := param.GetEnvelope(ts.transactor.Tip.ChainID())
	if txEnv == nil {
		return nil, fmt.Errorf("no transaction envelope or payload provided")
	}
	return ts.transactor.BroadcastTxSync(ctx, txEnv)
}

func (ts *transactServer) BroadcastTxAsync(ctx context.Context, param *TxEnvelopeParam) (*txs.Receipt, error) {
	txEnv := param.GetEnvelope(ts.transactor.Tip.ChainID())
	if txEnv == nil {
		return nil, fmt.Errorf("no transaction envelope or payload provided")
	}
	return ts.transactor.BroadcastTxAsync(txEnv)
}

func (ts *transactServer) SignTx(ctx context.Context, param *TxEnvelopeParam) (*TxEnvelope, error) {
	txEnv := param.GetEnvelope(ts.transactor.Tip.ChainID())
	if txEnv == nil {
		return nil, fmt.Errorf("no transaction envelope or payload provided")
	}
	txEnv, err := ts.transactor.SignTx(txEnv)
	if err != nil {
		return nil, err
	}
	return &TxEnvelope{
		Envelope: txEnv,
	}, nil
}

func (ts *transactServer) FormulateTx(ctx context.Context, param *PayloadParam) (*TxEnvelope, error) {
	txEnv := param.Envelope(ts.transactor.Tip.ChainID())
	if txEnv == nil {
		return nil, fmt.Errorf("no payload provided to FormulateTx")
	}
	return &TxEnvelope{
		Envelope: txEnv,
	}, nil
}

func (ts *transactServer) CallTxSync(ctx context.Context, param *payload.CallTx) (*exec.TxExecution, error) {
	return ts.BroadcastTxSync(ctx, txEnvelopeParam(param))
}

func (ts *transactServer) CallTxAsync(ctx context.Context, param *payload.CallTx) (*txs.Receipt, error) {
	return ts.BroadcastTxAsync(ctx, txEnvelopeParam(param))
}

func (ts *transactServer) CallTxSim(ctx context.Context, param *payload.CallTx) (*exec.TxExecution, error) {
	if param.Address == nil {
		return nil, fmt.Errorf("CallSim requires a non-nil address from which to retrieve code")
	}
	return ts.transactor.CallSim(param.Input.Address, *param.Address, param.Data)
}

func (ts *transactServer) CallCodeSim(ctx context.Context, param *CallCodeParam) (*exec.TxExecution, error) {
	return ts.transactor.CallCodeSim(param.FromAddress, param.Code, param.Data)
}

func (ts *transactServer) SendTxSync(ctx context.Context, param *payload.SendTx) (*exec.TxExecution, error) {
	return ts.BroadcastTxSync(ctx, txEnvelopeParam(param))
}

func (ts *transactServer) SendTxAsync(ctx context.Context, param *payload.SendTx) (*txs.Receipt, error) {
	return ts.BroadcastTxAsync(ctx, txEnvelopeParam(param))
}

func (ts *transactServer) NameTxSync(ctx context.Context, param *payload.NameTx) (*exec.TxExecution, error) {
	return ts.BroadcastTxSync(ctx, txEnvelopeParam(param))
}

func (ts *transactServer) NameTxAsync(ctx context.Context, param *payload.NameTx) (*txs.Receipt, error) {
	return ts.BroadcastTxAsync(ctx, txEnvelopeParam(param))
}

func (te *TxEnvelopeParam) GetEnvelope(chainID string) *txs.Envelope {
	if te == nil {
		return nil
	}
	if te.Envelope != nil {
		return te.Envelope
	}
	if te.Payload != nil {
		return te.Payload.Envelope(chainID)
	}
	return nil
}

func (pp *PayloadParam) Envelope(chainID string) *txs.Envelope {
	if pp.CallTx != nil {
		return txs.Enclose(chainID, pp.CallTx)
	}
	if pp.SendTx != nil {
		return txs.Enclose(chainID, pp.SendTx)
	}
	if pp.NameTx != nil {
		return txs.Enclose(chainID, pp.NameTx)
	}
	return nil
}

func txEnvelopeParam(pl payload.Payload) *TxEnvelopeParam {
	switch tx := pl.(type) {
	case *payload.CallTx:
		return &TxEnvelopeParam{
			Payload: &PayloadParam{
				CallTx: tx,
			},
		}
	case *payload.SendTx:
		return &TxEnvelopeParam{
			Payload: &PayloadParam{
				SendTx: tx,
			},
		}
	case *payload.NameTx:
		return &TxEnvelopeParam{
			Payload: &PayloadParam{
				NameTx: tx,
			},
		}
	}
	return nil
}