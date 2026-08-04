	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	arn := ackv1alpha1.AWSResourceName(databaseARN(ko))
	ko.Status.ACKResourceMetadata.ARN = &arn
	ko.Spec.Tags, err = rm.getTags(ctx, string(*ko.Status.ACKResourceMetadata.ARN))
