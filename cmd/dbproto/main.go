package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Malpizarr/dbproto/pkg/data"
	"github.com/Malpizarr/dbproto/pkg/exports"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func main() {
	envPath := "../../.env"
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	rootCmd := &cobra.Command{
		Use:   "dbproto",
		Short: "dbproto is a CLI for database interactions",
		Long:  `dbproto is a CLI that allows interactive database interactions.`,
	}

	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newExportCmd())

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to dbproto CLI. Type 'exit' to quit.")
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "exit" {
			color.Yellow("Exiting dbproto CLI.")
			break
		}

		args := strings.Fields(input)
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			color.Red("Error: %v", err)
		}
	}
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [database] [table]",
		Short: "List databases or tables within a database",
		Long:  `List all databases in the server or tables within a specific database.`,
		Run:   listFunc,
	}
	return cmd
}

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [database] [table] [filename]",
		Short: "Export records of a table to an XML file",
		Long:  `Export all records from a specified table in a database to an XML file.`,
		Run:   exportFunc,
	}
	return cmd
}

func exportFunc(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		fmt.Println("Usage: export [database] [table] [filename]")
		return
	}
	databaseName, tableName, filename := args[0], args[1], args[2]

	server := data.NewServer()
	if err := server.Initialize(); err != nil {
		color.Red("Failed to initialize server: %v", err)
		return
	}

	database, exists := server.Databases[databaseName]
	if !exists {
		color.Red("Database %s does not exist", databaseName)
		return
	}

	table, exists := database.Tables[tableName]
	if !exists {
		color.Red("Table %s does not exist", tableName)
		return
	}

	records, err := table.SelectAll()
	if err != nil {
		color.Red("Error retrieving records from table %s: %v", tableName, err)
		return
	}

	if err := exports.ExportRecordsToXML(records, filename); err != nil {
		color.Red("Error exporting records to XML: %v", err)
		return
	}

	color.Green("Records were successfully exported to %s", filename)
}

func listFunc(cmd *cobra.Command, args []string) {
	server := data.NewServer()
	if err := server.Initialize(); err != nil {
		color.Red("Failed to initialize server: %v", err)
		return
	}

	if len(args) == 0 {
		databases := server.ListDatabases()
		color.Green("Databases: %v", databases)
		return
	}

	databaseName := args[0]
	database, exists := server.Databases[databaseName]
	if !exists {
		color.Yellow("Database %s does not exist", databaseName)
		return
	}

	if len(args) == 1 {
		tables, err := database.ListTables()
		if err != nil {
			color.Red("Error listing tables in database %s: %v", databaseName, err)
			return
		}
		color.Cyan("Tables in %s: %v", databaseName, tables)
		return
	}

	tableName := args[1]
	table, exists := database.Tables[tableName]
	if !exists {
		color.Yellow("Table %s does not exist in database %s", tableName, databaseName)
		return
	}

	records, err := table.SelectAll()
	if err != nil {
		color.Red("Error retrieving records from table %s: %v", tableName, err)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	defer w.Flush()
	color.Magenta("Records in %s.%s:", databaseName, tableName)
	fmt.Fprintf(w, "Key\tValue\t\n")
	for i, record := range records {
		for key, val := range record.Fields {
			fmt.Fprintf(w, "%s\t%v\t\n", color.YellowString(key), formatProtoValue(val))
		}
		if i < len(records)-1 {
			fmt.Fprintln(w, "\t")
		}
	}
}

func formatProtoValue(val *structpb.Value) string {
	switch x := val.Kind.(type) {
	case *structpb.Value_StringValue:
		return fmt.Sprintf("\"%s\"", x.StringValue)
	case *structpb.Value_NumberValue:
		if float64(int(x.NumberValue)) == x.NumberValue {
			return fmt.Sprintf("%d", int(x.NumberValue))
		}
		return fmt.Sprintf("%.3f", x.NumberValue)
	case *structpb.Value_BoolValue:
		return fmt.Sprintf("%t", x.BoolValue)
	default:
		return fmt.Sprintf("%v", val)
	}
}
