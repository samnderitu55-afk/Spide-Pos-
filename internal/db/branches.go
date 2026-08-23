package db

import (
    "database/sql"
)

type Branch struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Location string `json:"location"`
    Phone    string `json:"phone"`
    IsActive bool   `json:"is_active"`
}

func GetBranches(db *sql.DB) ([]Branch, error) {
    query := `
        SELECT id, name, COALESCE(location, '') as location, 
               COALESCE(phone, '') as phone, is_active
        FROM branches
        WHERE is_active = 1
        ORDER BY name
    `
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var branches []Branch
    for rows.Next() {
        var b Branch
        err := rows.Scan(&b.ID, &b.Name, &b.Location, &b.Phone, &b.IsActive)
        if err != nil {
            return nil, err
        }
        branches = append(branches, b)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return branches, nil
}
