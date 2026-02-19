package config

import "fmt"

func MigrationTable() error {
	db, err := Dbconnection()
	if err != nil {
		return fmt.Errorf("database Connection Error..!%v", err)
	}

	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		return fmt.Errorf("failed to enable uuid-ossp: %v", err)
	}

	createUserQuery :=
		`CREATE TABLE IF NOT EXISTS users (
    		id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email 		VARCHAR(255),
    		username 	VARCHAR(255) ,
			password 	VARCHAR(255),
    		role 		VARCHAR(255) ,
    		created_at 	TIMESTAMP DEFAULT NOW(),
    		updated_at 	TIMESTAMP DEFAULT NOW()
	);`

	_, err = db.Exec(createUserQuery)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}


	createTaskQuery :=
		`CREATE TABLE IF NOT EXISTS tasks (
    		id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			title 		VARCHAR(255),
    		description TEXT ,
    		status 		VARCHAR(255) ,
    		user_id 	VARCHAR(255),
    		created_at 	TIMESTAMP DEFAULT NOW(),
    		updated_at 	TIMESTAMP DEFAULT NOW()
	);`

	_, err = db.Exec(createTaskQuery)
	if err != nil {
		return fmt.Errorf("error creating tasks table: %v", err)
	}

	return nil
}
