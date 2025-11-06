package pir

// Database of bytes
type Database struct {
	Num_records int    // number of records
	Record_len  int    // length of each record, in bytes
	Data        []byte 
}

func RandDatabase(num_records, record_len int) *Database {
	db := new(Database)
	db.Num_records = num_records
	db.Record_len = record_len
	db.Data = RandByteVec(num_records * record_len) 
	return db
}

func (db *Database) Read(index int) []byte {
	if index >= db.Num_records {
		panic("Read: reading out-of-bounds index")
	}

	return db.Data[index*db.Record_len : (index+1)*db.Record_len]
}
