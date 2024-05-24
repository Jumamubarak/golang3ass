package data

import "database/sql"

type Models struct {
	UserInfo UserInfoModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		UserInfo: UserInfoModel{db: db},
	}
}
