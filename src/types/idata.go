package types

//type DataType int
//
//const (
//	BlockData       DataType = 1
//	ReceiptData     DataType = 2
//	TransactionData DataType = 3
//)
//
//type Data interface {
//	FromProtobuf(dataType string) error
//	FromGoType(dataType string) error
//}
//
//func NewStore(t DataType) Data {
//	switch t {
//	case BlockData:
//		return models.BlockProtobufFromGoType
//	case ReceiptData:
//		return newDiskStorage( /*...*/)
//	case TransactionData:
//		return newDiskStorage( /*...*/)
//	default:
//		return newTempStorage( /*...*/)
//	}
//}
