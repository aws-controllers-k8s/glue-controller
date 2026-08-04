    databaseInput, err := rm.buildDatabaseInput(desired)
    if err != nil {
        return nil, err
    }
    input.DatabaseInput = databaseInput
