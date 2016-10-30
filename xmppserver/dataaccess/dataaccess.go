package dataaccess

import (
  "database/sql"
  "fmt"
  "os"
  "sync"
  "strconv"
  "xmppserver/logger"

  _ "github.com/mattn/go-sqlite3"
)

var dataaccess *DataAccess
var once sync.Once

type DataAccess struct {
  dbname string
  log logger.Logger
}

type user struct {
  fname string
  lname string
  login string
  pass string
  created_at string
  id int
}

func (dataaccess DataAccess) connect() (*sql.DB, bool) {
  dataaccess.log.Debug("Connecting to database")
  db, err := sql.Open("sqlite3", dataaccess.dbname)
  return db, !dataaccess.checkErr("Error connecting to database", err)
}

func (dataaccess DataAccess) LookupUserByLogin(login string) (*user, bool) {
    db, success := dataaccess.connect()
    defer db.Close()
    if (success) {
      dataaccess.log.Debug("Looking up user with login: " + login)
      rows, err := db.Query("SELECT first_name, last_name, password from user where login = '" + login + "'")
      defer rows.Close()
      if (!dataaccess.checkErr("Error looking up user", err)) {
          for rows.Next() {
            u := user{}
            err = rows.Scan(&u.fname, &u.lname, &u.pass)
            if dataaccess.checkErr("Error scanning row", err) {
              return nil, false
            }
            return &u, true
          }
        }
        return nil, true
    }
    return nil, false
}

func (dataaccess DataAccess) CreateUser(fname string, lname string, login string, pass string) (string, bool) {
  dataaccess.log.Debug("Create User")
  db, success := dataaccess.connect()
  defer db.Close()
  if (success) {
    dataaccess.log.Debug("Creating user with login: " + login)
    stmt, err := db.Prepare("INSERT INTO user (first_name, last_name, login, password) VALUES (?,?,?,?)")
    defer stmt.Close()
    if !dataaccess.checkErr("Error preparing statement", err) {
      res, err := stmt.Exec(fname, lname, login, pass)
      if !dataaccess.checkErr("Error executing statment", err) {
        id, _ := res.LastInsertId()
        idstr := strconv.FormatInt(id, 10)
        dataaccess.log.Debug("Created user with id: " +  idstr)
        return idstr, true
      }
    }
  }
  return "-1", false
}

func (dataaccess DataAccess) UpdateUser(u *user) (bool) {
  db, success := dataaccess.connect()
  defer db.Close()
  if (success) {
    stmt, err := db.Prepare("UPDATE user SET first_name = ?, last_name = ?, password = ? where login = ?")
    defer stmt.Close()
    if !dataaccess.checkErr("Error preparing statement for user update", err) {
      res, err := stmt.Exec(u.fname, u.lname, u.pass, u.login)
      if !dataaccess.checkErr("Error updating user", err) {
        affect, _ := res.RowsAffected()
        if affect > 0 {
          dataaccess.log.Debug("Updated user with login: " +  u.login)
          return true
        }
      }
    }
  }
  return false
}

func (dataaccess DataAccess) CreateOrUpdateUser(fname string, lname string, login string, pass string) (bool) {
  u, success := dataaccess.LookupUserByLogin(login)
  if success && u != nil {
    u.login = login
    if lname != "" {
      u.lname = lname
    }
    if fname != "" {
      u.fname = fname
    }
    if pass != "" {
      u.pass = pass
    }
    return dataaccess.UpdateUser(u)
  } else if success {
    _, success = dataaccess.CreateUser(fname, lname, login, pass)
    return success
  }
  return false
}

func GetDataAccess() *DataAccess {
  newlog := logger.Logger{LogLevel: 0, TAG: "DataAccess"}
  once.Do(func() {
    dataaccess = &DataAccess{dbname: "developer.db", log: newlog}
    dataaccess.log.Init(os.Stdout)
  })
  return dataaccess
}

func (dataaccess DataAccess) checkErr(msg string, err error) (bool) {
  if err != nil {
    dataaccess.log.Error(fmt.Sprintf("%s %s", msg, err.Error()))
    return true
  } else {
    return false
  }
}