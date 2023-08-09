package models

import (
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty"`
	BuyerName           string             `bson:"buyerName"`           // 구매자명
	BuyerPhone          string             `bson:"buyerPhone"`          // 구매자 연락처
	PaymentMethod       string             `bson:"paymentMethod"`       // 결제 수단
	ReceiverName        string             `bson:"receiverName"`        // 수취인명
	ReceiverPhone       string             `bson:"receiverPhone"`       // 수취인 연락처
	ReceiverAddress     string             `bson:"receiverAddress"`     // 배송주소
	ShippingFee         int                `bson:"shippingFee"`         // 배송비
	ShippingCompanyCode string             `bson:"shippingCompanyCode"` // 택배사
	ShippingCompanyName string             `bson:"shippingCompanyName"` // 택배사명
	InvoiceNumber       string             `bson:"invoiceNumber"`       // 운송장 번호
	OrderStatus         string             `bson:"orderStatus"`         // 주문 상태
}
