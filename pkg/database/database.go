package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db   *sql.DB
	path string
}

func NewDatabase(engine string, path string) (*Database, error) {
	db, err := sql.Open(engine, path)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	database := &Database{
		path: path,
		db:   db,
	}
	database.Init()
	return database, nil
}
func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) GetCompanies() ([]Company, error) {
	rows, err := d.db.Query("SELECT * FROM companies")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var company Company
		err := rows.Scan(&company.Name, &company.PPH)
		if err != nil {
			log.Fatal(err)
		}
		companies = append(companies, company)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return companies, nil
}

func (d *Database) InsertCompany(company Company) error {
	_, err := d.db.Exec("INSERT INTO companies VALUES (?, ?)", company.Name, company.PPH)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (d *Database) DeleteCompany(company Company) error {
	_, err := d.db.Exec("DELETE FROM companies WHERE name = ?", company.Name)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (d *Database) GetCompany(name string) (*Company, error) {
	rows, err := d.db.Query("SELECT * FROM companies WHERE name = ? COLLATE NOCASE", name)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var company Company
	for rows.Next() {
		err := rows.Scan(&company.Name, &company.PPH, &company.Currency)
		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	if company.Name == "" {
		return nil, nil
	}
	return &company, nil
}

func (d *Database) GetAllCompanies() (*[]Company, error) {
	rows, err := d.db.Query("SELECT * FROM companies")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var company Company
		err := rows.Scan(&company.Name, &company.PPH)
		if err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &companies, nil
}

func (d *Database) UpdateCompany(company Company) error {
	_, err := d.db.Exec("UPDATE companies SET pph = ? WHERE name = ?", company.PPH, company.Name)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (d *Database) InsertNewDay(day Day) error {
	date := day.Start.Format("2006-01-02")
	startTime := day.Start.Format("15:04:05")
	var endTime string
	if day.End != nil {
		endTime = day.End.Format("15:04:05")
	}
	_, err := d.db.Exec("INSERT INTO days VALUES (?, ?, ?, ?)", day.Company.Name, date, startTime, endTime)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) GetDay(datetime time.Time, company Company) (*Day, error) {
	date := datetime.Format("2006-01-02")
	rows, err := d.db.Query("SELECT company, start_time, end_time FROM days WHERE company = ? AND date = ?", company.Name, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var day Day
	dayFound := false

	for rows.Next() {
		dayFound = true
		var startTime, endTime string
		err := rows.Scan(&day.Company.Name, &startTime, &endTime)
		if err != nil {
			return nil, err
		}
		if startTime != "" {
			startDateTime := startTime + " " + date
			start, err := time.Parse("15:04:05 2006-01-02", startDateTime)
			if err != nil {
				return nil, err
			}
			day.Start = &start
		}
		if endTime != "" {
			endDateTime := endTime + " " + date
			end, err := time.Parse("15:04:05 2006-01-02", endDateTime)
			if err != nil {
				return nil, err
			}
			day.End = &end
		}
		company, err := d.GetCompany(day.Company.Name)
		if err != nil {
			return nil, err
		}
		day.Company = *company
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if !dayFound {
		return nil, fmt.Errorf("%s wasn't recorded", datetime.Format("2006-01-02"))
	}

	return &day, nil
}

func (d *Database) UpdateDay(day Day) error {
	date := day.Start.Format("2006-01-02")
	var startTime string
	if day.Start != nil {
		startTime = day.Start.Format("15:04:05")
	}
	var endTime string
	if day.End != nil {
		endTime = day.End.Format("15:04:05")
	}
	_, err := d.db.Exec("UPDATE days SET start_time = ?, end_time = ? WHERE company = ? AND date = ?", startTime, endTime, day.Company.Name, date)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) GetAllDays(company Company) (*[]Day, error) {
	rows, err := d.db.Query("SELECT date, start_time, end_time FROM days WHERE company = ? ORDER BY date DESC", company.Name)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var days []Day
	for rows.Next() {
		var day Day
		var date, startTime, endTime string
		err := rows.Scan(&date, &startTime, &endTime)
		if err != nil {
			return nil, err
		}
		if startTime != "" {
			startDateTime := startTime + " " + date
			start, err := time.Parse("15:04:05 2006-01-02", startDateTime)
			if err != nil {
				return nil, err
			}
			day.Start = &start
		}
		if endTime != "" {
			endDateTime := endTime + " " + date
			end, err := time.Parse("15:04:05 2006-01-02", endDateTime)
			if err != nil {
				return nil, err
			}
			day.End = &end
		}
		company, err := d.GetCompany(company.Name)
		if err != nil {
			return nil, err
		}
		day.Company = *company
		days = append(days, day)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &days, nil
}

func (d *Database) Init() {
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS companies (name TEXT NOT NULL PRIMARY KEY, pph INTEGER NOT NULL, currency TEXT NOT NULL);
    CREATE TABLE IF NOT EXISTS days (company TEXT NOT NULL, date TEXT NOT NULL, start_time TEXT NOT NULL, end_time TEXT, PRIMARY KEY (company, date), FOREIGN KEY (company) REFERENCES companies(name));
    `
	_, err := d.db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}
}
