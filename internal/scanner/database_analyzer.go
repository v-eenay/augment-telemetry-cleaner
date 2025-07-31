package scanner

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"augment-telemetry-cleaner/internal/utils"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseAnalysisResult represents the result of analyzing VS Code's database
type DatabaseAnalysisResult struct {
	ExtensionEntries    []DatabaseEntry `json:"extension_entries"`
	TelemetryEntries    []DatabaseEntry `json:"telemetry_entries"`
	UsageEntries        []DatabaseEntry `json:"usage_entries"`
	ConfigEntries       []DatabaseEntry `json:"config_entries"`
	TotalEntries        int             `json:"total_entries"`
	HighRiskEntries     int             `json:"high_risk_entries"`
	DatabasePath        string          `json:"database_path"`
	ScanDuration        time.Duration   `json:"scan_duration"`
}

// DatabaseEntry represents an entry found in VS Code's database
type DatabaseEntry struct {
	Table           string        `json:"table"`
	Key             string        `json:"key"`
	Value           string        `json:"value"`
	ExtensionID     string        `json:"extension_id,omitempty"`
	Risk            TelemetryRisk `json:"risk"`
	Category        string        `json:"category"`
	Description     string        `json:"description"`
	Size            int64         `json:"size"`
	LastModified    time.Time     `json:"last_modified,omitempty"`
}

// DatabaseAnalyzer handles analysis of VS Code's SQLite database
type DatabaseAnalyzer struct {
	telemetryKeyPatterns map[string]TelemetryRisk
	extensionPatterns    map[string]TelemetryRisk
	tableAnalyzers       map[string]func(*sql.DB, *DatabaseAnalysisResult) error
}

// NewDatabaseAnalyzer creates a new database analyzer
func NewDatabaseAnalyzer() *DatabaseAnalyzer {
	analyzer := &DatabaseAnalyzer{
		tableAnalyzers: make(map[string]func(*sql.DB, *DatabaseAnalysisResult) error),
	}
	analyzer.initializeTelemetryKeyPatterns()
	analyzer.initializeExtensionPatterns()
	analyzer.initializeTableAnalyzers()
	return analyzer
}

// initializeTelemetryKeyPatterns sets up patterns for telemetry-related database keys
func (da *DatabaseAnalyzer) initializeTelemetryKeyPatterns() {
	da.telemetryKeyPatterns = map[string]TelemetryRisk{
		// Direct telemetry keys
		"telemetry":                    TelemetryRiskHigh,
		"analytics":                    TelemetryRiskHigh,
		"tracking":                     TelemetryRiskHigh,
		"usage":                        TelemetryRiskMedium,
		"metrics":                      TelemetryRiskMedium,
		"statistics":                   TelemetryRiskMedium,
		"performance":                  TelemetryRiskLow,
		
		// Machine/user identification
		"machineid":                    TelemetryRiskCritical,
		"deviceid":                     TelemetryRiskCritical,
		"sessionid":                    TelemetryRiskHigh,
		"userid":                       TelemetryRiskHigh,
		"installid":                    TelemetryRiskHigh,
		"hostname":                     TelemetryRiskHigh,
		
		// Extension-related
		"extension.telemetry":          TelemetryRiskHigh,
		"extension.analytics":          TelemetryRiskHigh,
		"extension.usage":              TelemetryRiskMedium,
		"extension.performance":        TelemetryRiskLow,
		
		// Activity tracking
		"lastused":                     TelemetryRiskLow,
		"activationcount":              TelemetryRiskLow,
		"commandhistory":               TelemetryRiskMedium,
		"searchhistory":                TelemetryRiskMedium,
		"recentfiles":                  TelemetryRiskLow,
		
		// Error and crash data
		"crashreport":                  TelemetryRiskMedium,
		"errorlog":                     TelemetryRiskMedium,
		"diagnostic":                   TelemetryRiskMedium,
		
		// Configuration and experiments
		"experiment":                   TelemetryRiskMedium,
		"feature.flag":                 TelemetryRiskLow,
		"survey":                       TelemetryRiskMedium,
		"feedback":                     TelemetryRiskLow,
	}
}

// initializeExtensionPatterns sets up patterns for extension-specific database entries
func (da *DatabaseAnalyzer) initializeExtensionPatterns() {
	da.extensionPatterns = map[string]TelemetryRisk{
		// Extension activation and usage
		"extension.activation":         TelemetryRiskMedium,
		"extension.deactivation":       TelemetryRiskMedium,
		"extension.usage.count":        TelemetryRiskMedium,
		"extension.command.usage":      TelemetryRiskMedium,
		"extension.error.count":        TelemetryRiskMedium,
		
		// Extension storage patterns
		"globalStorage":                TelemetryRiskMedium,
		"workspaceStorage":             TelemetryRiskLow,
		"memento":                      TelemetryRiskLow,
		
		// Extension configuration
		"extension.config":             TelemetryRiskLow,
		"extension.settings":           TelemetryRiskLow,
		"extension.preferences":        TelemetryRiskLow,
		
		// Extension update and management
		"extension.update.check":       TelemetryRiskMedium,
		"extension.install.source":     TelemetryRiskMedium,
		"extension.uninstall.reason":   TelemetryRiskMedium,
	}
}

// initializeTableAnalyzers sets up specialized analyzers for different database tables
func (da *DatabaseAnalyzer) initializeTableAnalyzers() {
	da.tableAnalyzers["ItemTable"] = da.analyzeItemTable
	da.tableAnalyzers["ExtensionTable"] = da.analyzeExtensionTable
	da.tableAnalyzers["StateTable"] = da.analyzeStateTable
}

// AnalyzeDatabase performs comprehensive analysis of VS Code's database
func (da *DatabaseAnalyzer) AnalyzeDatabase() (*DatabaseAnalysisResult, error) {
	startTime := time.Now()
	
	// Get database path
	dbPath, err := utils.GetDBPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	result := &DatabaseAnalysisResult{
		ExtensionEntries: make([]DatabaseEntry, 0),
		TelemetryEntries: make([]DatabaseEntry, 0),
		UsageEntries:     make([]DatabaseEntry, 0),
		ConfigEntries:    make([]DatabaseEntry, 0),
		DatabasePath:     dbPath,
	}

	// Open database connection
	db, err := da.openDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get list of tables
	tables, err := da.getDatabaseTables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get database tables: %w", err)
	}

	// Analyze each table
	for _, table := range tables {
		if analyzer, exists := da.tableAnalyzers[table]; exists {
			// Use specialized analyzer
			if err := analyzer(db, result); err != nil {
				// Continue with other tables even if one fails
				continue
			}
		} else {
			// Use generic analyzer
			if err := da.analyzeGenericTable(db, table, result); err != nil {
				// Continue with other tables even if one fails
				continue
			}
		}
	}

	// Calculate totals and statistics
	da.calculateTotals(result)
	result.ScanDuration = time.Since(startTime)

	return result, nil
}

// openDatabase opens a connection to the VS Code database
func (da *DatabaseAnalyzer) openDatabase(dbPath string) (*sql.DB, error) {
	// Open with read-only mode and timeout
	connectionString := fmt.Sprintf("%s?mode=ro&_timeout=30000", dbPath)
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// getDatabaseTables gets a list of all tables in the database
func (da *DatabaseAnalyzer) getDatabaseTables(db *sql.DB) ([]string, error) {
	query := "SELECT name FROM sqlite_master WHERE type='table'"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue // Skip tables we can't read
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// analyzeItemTable analyzes the ItemTable (main key-value storage)
func (da *DatabaseAnalyzer) analyzeItemTable(db *sql.DB, result *DatabaseAnalysisResult) error {
	query := "SELECT key, value FROM ItemTable"
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query ItemTable: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue // Skip rows we can't read
		}

		entry := da.analyzeKeyValue("ItemTable", key, value)
		if entry != nil {
			da.categorizeEntry(*entry, result)
		}
	}

	return nil
}

// analyzeExtensionTable analyzes extension-specific tables
func (da *DatabaseAnalyzer) analyzeExtensionTable(db *sql.DB, result *DatabaseAnalysisResult) error {
	// This is a placeholder - actual VS Code database schema may vary
	query := "SELECT * FROM ExtensionTable LIMIT 1000"
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query ExtensionTable: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Prepare scan destinations
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			continue // Skip rows we can't read
		}

		// Analyze each column
		for i, column := range columns {
			if values[i] != nil {
				valueStr := fmt.Sprintf("%v", values[i])
				entry := da.analyzeKeyValue("ExtensionTable", column, valueStr)
				if entry != nil {
					da.categorizeEntry(*entry, result)
				}
			}
		}
	}

	return nil
}

// analyzeStateTable analyzes state-related tables
func (da *DatabaseAnalyzer) analyzeStateTable(db *sql.DB, result *DatabaseAnalysisResult) error {
	query := "SELECT key, value FROM StateTable"
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query StateTable: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue // Skip rows we can't read
		}

		entry := da.analyzeKeyValue("StateTable", key, value)
		if entry != nil {
			da.categorizeEntry(*entry, result)
		}
	}

	return nil
}

// analyzeGenericTable analyzes tables with unknown structure
func (da *DatabaseAnalyzer) analyzeGenericTable(db *sql.DB, tableName string, result *DatabaseAnalysisResult) error {
	// Get table schema
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to get table info for %s: %w", tableName, err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}
		
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			continue
		}
		columns = append(columns, name)
	}

	if len(columns) == 0 {
		return nil // No columns to analyze
	}

	// Query table data (limit to prevent performance issues)
	dataQuery := fmt.Sprintf("SELECT * FROM %s LIMIT 1000", tableName)
	dataRows, err := db.Query(dataQuery)
	if err != nil {
		return fmt.Errorf("failed to query table %s: %w", tableName, err)
	}
	defer dataRows.Close()

	// Prepare scan destinations
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for dataRows.Next() {
		if err := dataRows.Scan(valuePtrs...); err != nil {
			continue // Skip rows we can't read
		}

		// Analyze each column value
		for i, column := range columns {
			if values[i] != nil {
				valueStr := fmt.Sprintf("%v", values[i])
				entry := da.analyzeKeyValue(tableName, column, valueStr)
				if entry != nil {
					da.categorizeEntry(*entry, result)
				}
			}
		}
	}

	return nil
}

// analyzeKeyValue analyzes a key-value pair for telemetry patterns
func (da *DatabaseAnalyzer) analyzeKeyValue(table, key, value string) *DatabaseEntry {
	lowerKey := strings.ToLower(key)
	lowerValue := strings.ToLower(value)
	
	// Check against telemetry key patterns
	risk := TelemetryRiskNone
	category := "Unknown"
	description := ""

	// Check telemetry patterns
	for pattern, patternRisk := range da.telemetryKeyPatterns {
		if strings.Contains(lowerKey, strings.ToLower(pattern)) ||
		   strings.Contains(lowerValue, strings.ToLower(pattern)) {
			if patternRisk > risk {
				risk = patternRisk
				category = "Telemetry"
				description = fmt.Sprintf("Contains telemetry pattern: %s", pattern)
			}
		}
	}

	// Check extension patterns
	for pattern, patternRisk := range da.extensionPatterns {
		if strings.Contains(lowerKey, strings.ToLower(pattern)) ||
		   strings.Contains(lowerValue, strings.ToLower(pattern)) {
			if patternRisk > risk {
				risk = patternRisk
				category = "Extension"
				description = fmt.Sprintf("Contains extension pattern: %s", pattern)
			}
		}
	}

	// Skip entries with no telemetry risk
	if risk == TelemetryRiskNone {
		return nil
	}

	// Extract extension ID if possible
	extensionID := da.extractExtensionID(key, value)

	return &DatabaseEntry{
		Table:       table,
		Key:         key,
		Value:       da.sanitizeValue(value),
		ExtensionID: extensionID,
		Risk:        risk,
		Category:    category,
		Description: description,
		Size:        int64(len(value)),
	}
}

// extractExtensionID attempts to extract extension ID from key or value
func (da *DatabaseAnalyzer) extractExtensionID(key, value string) string {
	// Look for extension ID patterns in key
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		if len(parts) >= 2 {
			// Check if it looks like publisher.extension format
			if len(parts[0]) > 0 && len(parts[1]) > 0 {
				return parts[0] + "." + parts[1]
			}
		}
	}

	// Look for extension ID patterns in value
	if strings.Contains(value, ".") && len(value) < 100 {
		// Simple heuristic for extension ID in value
		parts := strings.Split(value, ".")
		if len(parts) == 2 && len(parts[0]) > 2 && len(parts[1]) > 2 {
			return value
		}
	}

	return ""
}

// sanitizeValue sanitizes a database value for safe display
func (da *DatabaseAnalyzer) sanitizeValue(value string) string {
	// Truncate very long values
	if len(value) > 200 {
		return value[:200] + "... (truncated)"
	}

	// Mask potentially sensitive data
	lowerValue := strings.ToLower(value)
	if strings.Contains(lowerValue, "password") ||
	   strings.Contains(lowerValue, "token") ||
	   strings.Contains(lowerValue, "secret") ||
	   strings.Contains(lowerValue, "key") {
		return "[SENSITIVE DATA MASKED]"
	}

	return value
}

// categorizeEntry categorizes a database entry into the appropriate result category
func (da *DatabaseAnalyzer) categorizeEntry(entry DatabaseEntry, result *DatabaseAnalysisResult) {
	switch entry.Category {
	case "Telemetry":
		result.TelemetryEntries = append(result.TelemetryEntries, entry)
	case "Extension":
		result.ExtensionEntries = append(result.ExtensionEntries, entry)
	default:
		// Categorize based on risk and content
		if entry.Risk >= TelemetryRiskMedium {
			result.TelemetryEntries = append(result.TelemetryEntries, entry)
		} else if entry.ExtensionID != "" {
			result.ExtensionEntries = append(result.ExtensionEntries, entry)
		} else if strings.Contains(strings.ToLower(entry.Key), "usage") ||
				 strings.Contains(strings.ToLower(entry.Key), "activity") {
			result.UsageEntries = append(result.UsageEntries, entry)
		} else {
			result.ConfigEntries = append(result.ConfigEntries, entry)
		}
	}
}

// calculateTotals calculates summary statistics for the analysis result
func (da *DatabaseAnalyzer) calculateTotals(result *DatabaseAnalysisResult) {
	allEntries := [][]DatabaseEntry{
		result.ExtensionEntries,
		result.TelemetryEntries,
		result.UsageEntries,
		result.ConfigEntries,
	}

	for _, entries := range allEntries {
		result.TotalEntries += len(entries)
		for _, entry := range entries {
			if entry.Risk >= TelemetryRiskHigh {
				result.HighRiskEntries++
			}
		}
	}
}

// GetDatabaseSchema returns information about the database schema
func (da *DatabaseAnalyzer) GetDatabaseSchema() (map[string][]string, error) {
	dbPath, err := utils.GetDBPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	db, err := da.openDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	tables, err := da.getDatabaseTables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	schema := make(map[string][]string)
	for _, table := range tables {
		columns, err := da.getTableColumns(db, table)
		if err != nil {
			continue // Skip tables we can't analyze
		}
		schema[table] = columns
	}

	return schema, nil
}

// getTableColumns gets the column names for a specific table
func (da *DatabaseAnalyzer) getTableColumns(db *sql.DB, tableName string) ([]string, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get table info: %w", err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}
		
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			continue
		}
		columns = append(columns, name)
	}

	return columns, nil
}