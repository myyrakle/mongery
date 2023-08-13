package semiorder

// @Entity
type SemiOrder struct {
	ID            string `bson:"_id,omitempty"`
	BuyerName     string `bson:"buyerName"`     // 구매자명
	BuyerPhone    string `bson:"buyerPhone"`    // 구매자 연락처
	PaymentMethod string `bson:"paymentMethod"` // 결제 수단
	ReceiverPhone string `bson:"receiverPhone"` // 수취인 연락처
}
