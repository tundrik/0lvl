package inspector

import "github.com/romshark/jscan/v2"


const (
	Keys = 14
	ObjKeys = 17
	ItemKeys = 11
)

type value struct {
	Type jscan.ValueType
}

func createScheme() map[string]value {
	scheme := map[string]value{
		"track_number": {
			Type: jscan.ValueTypeString,
		},// order, item
		"name": {
			Type: jscan.ValueTypeString,
		},// delivery, item

        "order_uid": {
			Type: jscan.ValueTypeString,
		},  
		"entry": {
			Type: jscan.ValueTypeString,
		}, 
		"locale": {
			Type: jscan.ValueTypeString,
		}, 
		"internal_signature": {
			Type: jscan.ValueTypeString,
		}, 
		"customer_id": {
			Type: jscan.ValueTypeString,
		}, 
		"delivery_service": {
			Type: jscan.ValueTypeString,
		}, 
		"shardkey": {
			Type: jscan.ValueTypeString,
		}, 
		"sm_id": {
			Type: jscan.ValueTypeNumber,
		}, 
		"date_created": {
			Type: jscan.ValueTypeString,
		}, 
		"oof_shard": {
			Type: jscan.ValueTypeString,
		}, 

		"delivery": {
			Type: jscan.ValueTypeObject,
		}, 
		"phone": {
			Type: jscan.ValueTypeString,
		}, 
		"zip": {
			Type: jscan.ValueTypeString,
		}, 
		"city": {
			Type: jscan.ValueTypeString,
		}, 
		"address": {
			Type: jscan.ValueTypeString,
		}, 
		"region": {
			Type: jscan.ValueTypeString,
		}, 
		"email": {
			Type: jscan.ValueTypeString,
		}, 

		"payment": {
			Type: jscan.ValueTypeObject,
		}, 
		"transaction": {
			Type: jscan.ValueTypeString,
		}, 
		"request_id": {
			Type: jscan.ValueTypeString,
		}, 
		"currency": {
			Type: jscan.ValueTypeString,
		}, 
		"provider": {
			Type: jscan.ValueTypeString,
		}, 
		"amount": {
			Type: jscan.ValueTypeNumber,
		}, 
		"payment_dt": {
			Type: jscan.ValueTypeNumber,
		}, 
		"bank": {
			Type: jscan.ValueTypeString,
		}, 
		"delivery_cost": {
			Type: jscan.ValueTypeNumber,
		}, 
		"goods_total": {
			Type: jscan.ValueTypeNumber,
		}, 
		"custom_fee": {
			Type: jscan.ValueTypeNumber,
		}, 

		"items": {
			Type: jscan.ValueTypeArray,
		}, 
		"chrt_id": {
			Type: jscan.ValueTypeNumber,
		}, 
		"price": {
			Type: jscan.ValueTypeNumber,
		}, 
		"rid": {
			Type: jscan.ValueTypeString,
		}, 
		"sale": {
			Type: jscan.ValueTypeNumber,
		}, 
		"size": {
			Type: jscan.ValueTypeString,
		}, 
		"total_price": {
			Type: jscan.ValueTypeNumber,
		}, 
		"nm_id": {
			Type: jscan.ValueTypeNumber,
		}, 
		"brand": {
			Type: jscan.ValueTypeString,
		}, 
		"status": {
			Type: jscan.ValueTypeNumber,
		}, 

    }

	return scheme
}



