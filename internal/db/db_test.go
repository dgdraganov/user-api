package db_test

import (
	"context"
	"database/sql"

	"github.com/dgdraganov/user-api/internal/db"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Test struct {
	ID       uint `gorm:"primaryKey"`
	Username string
}

var _ = Describe("Database", func() {
	var (
		mock   sqlmock.Sqlmock
		mockDb *sql.DB
		err    error
		testDB *db.MySQL
	)

	BeforeEach(func() {
		mockDb, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())

		mock.ExpectQuery("SELECT VERSION()").
			WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.0"))

		dialector := mysql.New(mysql.Config{
			Conn:       mockDb,
			DriverName: "mysql",
		})

		gormDB, err := gorm.Open(dialector, &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		Expect(err).NotTo(HaveOccurred())

		testDB = &db.MySQL{
			DB: gormDB,
		}

	})

	AfterEach(func() {
		mock.ExpectClose()
		Expect(mockDb.Close()).To(Succeed())
	})

	Describe("MigrateTable", func() {
		var err error

		BeforeEach(func() {
			//SELECT SCHEMA_NAME from Information_schema.SCHEMATA where SCHEMA_NAME LIKE ? ORDER BY SCHEMA_NAME=? DESC,SCHEMA_NAME limit 1
			mock.ExpectQuery(`SELECT SCHEMA_NAME from Information_schema.SCHEMATA where*`).
				WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(0))

			mock.ExpectExec("CREATE TABLE `tests`*").
				WillReturnResult(sqlmock.NewResult(0, 1))
		})
		JustBeforeEach(func() {
			err = testDB.MigrateTable(&Test{})
		})
		It("should migrate the table successfully", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})
	})

	Describe("InsertToTable", func() {
		var err error
		BeforeEach(func() {
			mock.ExpectBegin()

			mock.ExpectExec("INSERT INTO `tests` \\(`username`,`id`\\) VALUES *").
				WithArgs("Alice", 1, "Bob", 2).
				WillReturnResult(sqlmock.NewResult(0, 2))

			mock.ExpectCommit()
		})

		JustBeforeEach(func() {
			err = testDB.InsertToTable(context.Background(), &[]Test{
				{ID: 1, Username: "Alice"},
				{ID: 2, Username: "Bob"},
			})
		})

		It("should save records without errors", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})
	})

	Describe("GetOneBy", func() {
		When("a record is found", func() {
			BeforeEach(func() {
				mock.ExpectQuery("SELECT \\* FROM `tests` WHERE username = \\? ORDER BY `tests`.`id` LIMIT \\?").
					WithArgs("Alice", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
						AddRow(1, "Alice"))
			})

			It("should return the correct record", func() {
				var result Test
				err := testDB.GetOneBy(context.Background(), "username", "Alice", &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ID).To(Equal(uint(1)))
				Expect(result.Username).To(Equal("Alice"))
				Expect(mock.ExpectationsWereMet()).To(Succeed())
			})
		})

		When("no record is found", func() {
			BeforeEach(func() {
				mock.ExpectQuery("SELECT \\* FROM `tests` WHERE username = \\? ORDER BY `tests`.`id` LIMIT \\?").
					WithArgs("Ghost", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			})

			It("should return ErrNotFound", func() {
				var result Test
				err := testDB.GetOneBy(context.Background(), "username", "Ghost", &result)
				Expect(err).To(Equal(db.ErrNotFound))
				Expect(mock.ExpectationsWereMet()).To(Succeed())
			})
		})
	})

	Describe("GetAllBy", func() {
		When("multiple records are found", func() {
			BeforeEach(func() {
				mock.ExpectQuery("SELECT \\* FROM `tests` WHERE username IN \\(\\?,\\?\\).*").
					WithArgs("Alice", "Bob").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
						AddRow(1, "Alice").
						AddRow(2, "Bob"))
			})

			It("should return all matching records", func() {
				var results []Test
				err := testDB.GetAllBy(context.Background(), "username", []string{"Alice", "Bob"}, &results)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(2))
				Expect(results[0].Username).To(Equal("Alice"))
				Expect(results[1].Username).To(Equal("Bob"))
				Expect(mock.ExpectationsWereMet()).To(Succeed())
			})
		})

		When("an error occurs during query", func() {
			BeforeEach(func() {
				mock.ExpectQuery("SELECT \\* FROM `tests` WHERE username.*").
					WithArgs("Invalid").
					WillReturnError(sql.ErrConnDone)
			})

			It("should return an error", func() {
				var results []Test
				err := testDB.GetAllBy(context.Background(), "username", "Invalid", &results)
				Expect(err).To(MatchError(ContainSubstring("getting records by")))
				Expect(mock.ExpectationsWereMet()).To(Succeed())
			})
		})
	})

})
