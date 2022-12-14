package handler

import (
	"data-service/src/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type TableInfo struct {
	utils.JsonStandard
	Id       uint16 `json:"data_id"`
	Name     string `json:"data_name"`
	Type     uint16 `json:"data_type"`
	Subtype1 uint16 `json:"data_subtype1"`
	Subtype2 uint16 `json:"data_subtype2"`
	Rate     uint16 `json:"data_rate"`
	Size     uint16 `json:"data_size"`
	Unit     string `json:"data_unit"`
	Notes    string `json:"data_notes"`
}

func ReadDataInfo(path string) ([]TableInfo, error) {
	var dataList []TableInfo
	var objmap map[string]json.RawMessage

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Failed to read data configuration file")
		return nil, err
	}

	err = json.Unmarshal(byteValue, &objmap)
	if err != nil {
		fmt.Println("Failed to unmarshal data configuration json file")
		return nil, err
	}
	// dataList = make([]TableInfo, len(objmap))

	for _, value := range objmap {
		var info TableInfo
		err = json.Unmarshal(value, &info)
		if err != nil {
			return nil, err
		}
		dataList = append(dataList, info)
	}
	return dataList, nil

}

func DatabaseGenerator(src uint8, path string) error {
	// -------------------- Connect database ----------------------------------
	var err error
	var db *sql.DB
	var dbName string
	if src == utils.SRC_GCC {
		db, err = sql.Open("mysql", fmt.Sprintf("%v:%v@(ground_db:3306)/%v", // "hms_db" is the database container's name in the docker-compose.yml
			os.Getenv("DB_USER_GROUND"),
			os.Getenv("DB_PASSWORD_GROUND"),
			os.Getenv("DB_NAME_GROUND")))
		dbName = "ground"

	} else if src == utils.SRC_HMS {
		db, err = sql.Open("mysql", fmt.Sprintf("%v:%v@(habitat_db:3306)/%v", // "hms_db" is the database container's name in the docker-compose.yml
			os.Getenv("DB_USER_HABITAT"),
			os.Getenv("DB_PASSWORD_HABITAT"),
			os.Getenv("DB_NAME_HABITAT")))
		dbName = "habitat"
	}
	if err != nil {
		fmt.Println(err)
		return err
	}

	// wait database container to start
	for {
		err := db.Ping()
		if err == nil {
			break
		}
		fmt.Println(err)
		time.Sleep(1 * time.Second)
	}

	// ----------------------- Create table ------------------------------
	tableName := "info0"
	drop := fmt.Sprintf(`DROP TABLE IF EXISTS %s.%s`, dbName, tableName)
	_, err = db.Exec(drop)
	if err != nil {
		fmt.Println(err)
	}
	action := fmt.Sprintf(`CREATE TABLE %s.%s (
            data_id INT(16) UNSIGNED NOT NULL,
            data_name VARCHAR(128) NULL,
            data_type INT(8) UNSIGNED NOT NULL,
            data_subtype1 INT(8) UNSIGNED NULL,
            data_subtype2 INT(8) UNSIGNED NULL,
            data_rate INT(16) UNSIGNED NULL,
            data_size INT(16) UNSIGNED NULL,
            data_unit VARCHAR(45) NULL,
            data_notes VARCHAR(128) NULL,
            PRIMARY KEY (data_id),
            UNIQUE INDEX data_id_UNIQUE (data_id ASC) VISIBLE);`, dbName, tableName)

	_, err = db.Exec(action)
	if err != nil {
		fmt.Println(err)
	}

	// ---------------------- Read json file

	dataList, err := ReadDataInfo(path)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// ---------------------- Insert info

	for _, info := range dataList {
		act := fmt.Sprintf(`INSERT INTO %s.%s (data_id, data_name, data_type, data_subtype1, data_subtype2, data_rate, data_size) VALUES
				("%d", "%s", "%d", "%d","%d","%d", "%d");`, dbName, tableName, info.Id, info.Name, info.Type, info.Subtype1, info.Subtype2, info.Rate, info.Size)
		_, err = db.Exec(act)
		if err != nil {
			fmt.Println(err)
		}
	}

	// ---------------------- Create data table
	for _, info := range dataList {
		tableName = fmt.Sprintf(`record%d`, info.Id)
		drop := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tableName)
		db.Exec(drop)
		act := fmt.Sprintf("create table `%s` (", tableName)
		act = act + "simulink_time int unsigned NOT NULL,"
		act = act + "physical_time_s bigint unsigned NOT NULL,"
		act = act + "physical_time_d bigint unsigned NOT NULL,"
		for i := 0; i != int(info.Size); i++ {
			act = act + fmt.Sprintf("value%d float,", i)
		}
		act = act + "primary key (simulink_time), UNIQUE KEY simulink_time (simulink_time)"
		act = act + ")ENGINE=InnoDB"

		_, err = db.Exec(act)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("Database has been initialized")
	return nil

}
