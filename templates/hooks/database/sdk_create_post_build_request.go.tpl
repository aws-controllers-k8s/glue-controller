    input.CatalogId = desired.ko.Spec.CatalogID
    if desired.ko.Spec.Tags != nil {
        input.Tags = aws.ToStringMap(desired.ko.Spec.Tags)
    }
