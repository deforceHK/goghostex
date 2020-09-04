package goghostex

type LoanStatus int

const (
	LOAN_UNFINISH LoanStatus = iota
	LOAN_PART_FINISH
	LOAN_FINISH
	_
	_
	_
	LOAN_FAIL
	LOAN_REPAY
)

type MarginAccount struct {
	Pair       Pair
	SubAccount map[string]MarginSubAccount

	LiquidationPrice float64 //预计爆仓价格
	RiskRate         float64
	MarginRatio      float64
}

type MarginSubAccount struct {
	Currency Currency
	// Amount = AmountNet + AmountLoaned = AmountAvail + AmountFrozen
	Amount float64
	// 可用额度
	AmountAvail float64
	// 冻结额度
	AmountFrozen float64
	// 净值额度
	AmountNet float64
	// 已借贷额度
	AmountLoaned float64
	// 当前借贷费用
	LoaningFee float64
}

type Loan struct {
	Pair              Pair       // The loan currency pair
	Currency          Currency   // Currency
	Amount            float64    // Loan amount
	AmountLoaned      float64    // Loaned amount
	AmountInterest    float64    // Loan interest
	Status            LoanStatus // The loan record status
	LoanId            string     // Remote loan record id
	LoanTimestamp     int64      // Loan timestamp
	LoanDate          string     // Loan date
	RepayId           string     // Remote loan record repay id
	RepayTimestamp    int64      // Repay Timestamp
	RepayDate         string     // Repay Date
	RepayDeadlineDate string     // Repay Deadline Date
}
