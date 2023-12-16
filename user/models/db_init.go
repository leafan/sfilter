package models

import (
	"sfilter/user/config"
	"sfilter/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

func initTables(db *mongo.Client) {
	utils.DoInitTable(config.DatabaseName, config.UserTableName, UserIndexModel, db)
	utils.DoInitTable(config.DatabaseName, config.VerifyCodeTableName, VerifyCodeIndexModel, db)
	utils.DoInitTable(config.DatabaseName, config.LoginHistoryTableName, LoginHistoryIndexModel, db)
}

func checkOrCreatAdmin() {
	admin := "admin"
	_, err := GetUser(admin)

	if err != nil {
		referCode := "deepeye"

		user := &User{
			BasicInfo: BasicInfo{
				Username: admin,
				Email:    "market@deepeye.cc",
				Nickname: "Deepeye_Admin",
				Passwd:   "1",
			},
			ReferInfo: ReferInfo{
				ReferCode: &referCode,
			},
			RoleInfo: RoleInfo{
				Role: 9999,
			},
		}

		err = CreateUser(user)
		if err != nil {
			utils.Fatalf("[ checkOrCreatAdmin ] err: %v", err)
		}
	}

}
