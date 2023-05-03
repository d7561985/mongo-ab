package config

type Mongo struct {
	Addr string
	DB   string

	ShardNum int

	Collections struct {
		// for increment operation
		Balance string

		// for insert operation
		Journal string
	}

	// Note! Only single compression!
	Compression struct {
		// zlib, zstd, snappy
		Type string

		// zlib: max 9
		// zstd: max 20
		Level int
	}

	WriteConcert struct {
		Enabled bool
		Journal bool
		W       int
	}
}

type Postgres struct {
	Addr string
	//DB   string

	//Table struct {
	//	// for increment operation
	//	Balance string
	//
	//	// for insert operation
	//	Journal string
	//}

}
