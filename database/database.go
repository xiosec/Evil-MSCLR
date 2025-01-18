package database

import (
	"database/sql"
	"fmt"
)

var DB *sql.DB

func Init(host, user, pass string, port int) error {
	conn, err := sql.Open("mssql", fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d", host, user, pass, port))
	if err != nil {
		return err
	}
	DB = conn
	return nil
}

func AssemblyList() ([]Assembly, error) {
	stmt, err := DB.Prepare("select name, assembly_id, clr_name, permission_set_desc, is_user_defined  from sys.assemblies")
	if err != nil {
		return []Assembly{}, err
	}
	row, err := stmt.Query()
	if err != nil {
		return []Assembly{}, err
	}

	assembly := []Assembly{}
	for row.Next() {
		tmp := Assembly{}
		row.Scan(&tmp.Name, &tmp.Assembly_id, &tmp.CLR_name, &tmp.Permission_set_desc, &tmp.Is_user_defined)
		assembly = append(assembly, tmp)
	}

	return assembly, nil
}

func CLRStatus() (bool, error) {
	stmt, err := DB.Prepare("SELECT value FROM sys.configurations WHERE name = 'clr enabled'")
	if err != nil {
		return false, err
	}

	var status bool
	err = stmt.QueryRow().Scan(&status)

	return status, err
}

func ChangeCLR(status int8) error {
	_, err := DB.Exec("sp_configure 'clr enabled', ?; RECONFIGURE", status)

	return err
}

func SetTRUSTWORTHY(status bool) error {
	arg := "OFF"
	if status {
		arg = "ON"
	}
	_, err := DB.Exec("ALTER DATABASE master SET TRUSTWORTHY", arg)

	return err
}

func LoadAssembly(name, assembly string) error {
	_, err := DB.Exec(fmt.Sprintf("CREATE ASSEMBLY [%s] AUTHORIZATION [dbo] FROM %s WITH PERMISSION_SET = UNSAFE;", name, assembly))

	return err
}

func CreateProcedure(procedureName, assemblyName string) error {
	_, err := DB.Exec(fmt.Sprintf("CREATE PROCEDURE [dbo].[%s] @procedure NVARCHAR (MAX) NULL, @arg NVARCHAR (MAX) NULL AS EXTERNAL NAME [%s].[StoredProcedures].[SqlStoredProcedure]", procedureName, assemblyName))

	return err
}

func ExecFunction(function, arg string) (string, error) {
	r := DB.QueryRow(fmt.Sprintf("exec SqlStoredProcedure '%s','%s'", function, arg))
	var out string
	err := r.Scan(&out)
	return out, err
}

func DropAssembly(assemblyName string) error {
	_, err := DB.Exec(fmt.Sprintf("drop assembly %s", assemblyName))

	return err
}

func DropProcedure(procedureName string) error {
	_, err := DB.Exec(fmt.Sprintf("drop procedure %s", procedureName))

	return err
}
