package converter

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/DelineaXPM/delinea-netconfig/pkg/types"
)

// CSVConverter converts network entries to CSV format
type CSVConverter struct{}

// Convert converts network entries to CSV
func (c *CSVConverter) Convert(entries []types.NetworkEntry) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// Write header
	header := []string{"direction", "service", "region", "type", "value", "protocol", "ports", "description", "redundancy"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write entries
	for _, entry := range entries {
		// Create one row per value
		for _, value := range entry.Values {
			row := []string{
				entry.Direction,
				entry.Service,
				entry.Region,
				entry.Type,
				value,
				entry.Protocol,
				formatPorts(entry.Ports),
				entry.Description,
				entry.Redundancy,
			}

			if err := writer.Write(row); err != nil {
				return nil, fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// Name returns the name of the converter
func (c *CSVConverter) Name() string {
	return "CSV"
}

// FileExtension returns the file extension for CSV files
func (c *CSVConverter) FileExtension() string {
	return "csv"
}

// formatPorts converts a slice of ports to a comma-separated string
func formatPorts(ports []int) string {
	if len(ports) == 0 {
		return ""
	}

	portStrs := make([]string, len(ports))
	for i, port := range ports {
		portStrs[i] = strconv.Itoa(port)
	}
	return strings.Join(portStrs, ";")
}
