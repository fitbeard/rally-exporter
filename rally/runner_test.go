package rally

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RunnerSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock
}

func (s *RunnerSuite) SetupSuite() {
	db, mock, err := sqlmock.New()
	require.NoError(s.T(), err)

	s.DB, err = gorm.Open("sqlite3", db)
	require.NoError(s.T(), err)

	s.DB.LogMode(true)
	s.mock = mock
}

func (s *RunnerSuite) TestGetLatestTaskWithNoRecords() {
	s.mock.
		ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tasks" WHERE ("tasks"."status" NOT IN (?)) ORDER BY "tasks"."id" DESC LIMIT 1`)).
		WithArgs("running").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status"}),
		)

	_, err := getLatestTask(s.DB)
	require.Equal(s.T(), err, gorm.ErrRecordNotFound)
}

func (s *RunnerSuite) TestGetLatestTask() {
	s.mock.
		ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tasks" WHERE ("tasks"."status" NOT IN (?)) ORDER BY "tasks"."id" DESC LIMIT 1`)).
		WithArgs("running").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status"}).
				AddRow(1, "finished"),
		)

	task, err := getLatestTask(s.DB)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "finished", task.Status)
}
func TestRunnerSuite(t *testing.T) {
	suite.Run(t, new(RunnerSuite))
}
