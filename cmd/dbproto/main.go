package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Malpizarr/dbproto/pkg/data"
	"github.com/Malpizarr/dbproto/pkg/exports"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func main() {
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
	var format string
	cmd := &cobra.Command{
		Use:   "export [database] [table] [filename]",
		Short: "Export records of a table to a specified format",
		Long:  `Export all records from a specified table in a database to a specified format (e.g., CSV, XML).`,
		Run:   exportFunc,
	}
	cmd.Flags().StringVarP(&format, "format", "f", "csv", "Format to export (csv, xml)")
	return cmd
}

func exportFunc(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		fmt.Println("Usage: export [database] [table] [filename] --format=[csv|xml]")
		return
	}
	databaseName, tableName, filename := args[0], args[1], args[2]

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		color.Red("Error retrieving format flag: %v", err)
		return
	}

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

	switch format {
	case "csv":
		if err := exports.ExportRecordsToCSV(records, filename); err != nil {
			color.Red("Error exporting records to CSV: %v", err)
			return
		}
	case "xml":
		if err := exports.ExportRecordsToXML(records, filename); err != nil {
			color.Red("Error exporting records to XML: %v", err)
			return
		}
	default:
		color.Red("Unsupported format %s", format)
		return
	}

	color.Green("Records were successfully exported to %s in %s format", filename, format)
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
