# mongery

![](https://img.shields.io/badge/language-Go-00ADD8) ![](https://img.shields.io/badge/version-0.5.0-brightgreen) [![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

[document](https://pkg.go.dev/github.com/myyrakle/mongery)

## install

```
go install github.com/myyrakle/mongery@v0.5.0
```

## confiuration

The `.mongery.yaml` file must exist in the project root path.

Here is an example of a config file.

```
basedir: example
output-suffix: "_field.go"
```

It means that all files in the example directory will be read, and the output file will be created with the name "\*\_field.go".

## How to use?

Usage is very simple. Just run the following command in your project root path:

```
mongery
```

mongery only generates structures with `// @Entity` comments. It reads the bson tag value and creates a list of constants.

If you have a struct like

```
// @Entity
type Order struct {
	ID                  string `bson:"_id,omitempty"`
	BuyerName           string `bson:"buyerName"`           // 구매자명
	BuyerPhone          string `bson:"buyerPhone"`          // 구매자 연락처
	PaymentMethod       string `bson:"paymentMethod"`       // 결제 수단
	ReceiverName        string `bson:"receiverName"`        // 수취인명
	ReceiverPhone       string `bson:"receiverPhone"`       // 수취인 연락처
	ReceiverAddress     string `bson:"receiverAddress"`     // 배송주소
	ShippingFee         int    `bson:"shippingFee"`         // 배송비
	ShippingCompanyCode string `bson:"shippingCompanyCode"` // 택배사
	ShippingCompanyName string `bson:"shippingCompanyName"` // 택배사명
	InvoiceNumber       string `bson:"invoiceNumber"`       // 운송장 번호
	OrderStatus         string `bson:"orderStatus"`         // 주문 상태
}
```

mongery produces a list of constants like this:

```
const Order_ID = "_id"
const Order_BuyerName = "buyerName"
const Order_BuyerPhone = "buyerPhone"
const Order_PaymentMethod = "paymentMethod"
const Order_ReceiverName = "receiverName"
const Order_ReceiverPhone = "receiverPhone"
const Order_ReceiverAddress = "receiverAddress"
const Order_ShippingFee = "shippingFee"
const Order_ShippingCompanyCode = "shippingCompanyCode"
const Order_ShippingCompanyName = "shippingCompanyName"
const Order_InvoiceNumber = "invoiceNumber"
const Order_OrderStatus = "orderStatus"
```

## Features - Slice 

```yaml
features:
  - SLICE
```
If you enable Slice among the features flags, it creates a basic boilerplate for Slice.


It is as follows:
```go
type PersonSoManies []PersonSoMany

func (t PersonSoManies) Len() int {
	return len(t)
}

func (t PersonSoManies) IsEmpty() bool {
	return len(t) == 0
}

func (t PersonSoManies) Append(v PersonSoMany) PersonSoManies {
	t = append(t, v)
	return s
}

func (t PersonSoManies) First() PersonSoMany {
	if t.IsEmpty() {
		return PersonSoMany{}
	}
	return t[0]
}

func (t PersonSoManies) Last() PersonSoMany {
	if t.IsEmpty() {
		return PersonSoMany{}
	}
	return t[len(t)-1]
}
```
