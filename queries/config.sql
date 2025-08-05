-- name: GetPrinterConfig :one
SELECT value from config where key = 'thermal-printer' LIMIT 1;
