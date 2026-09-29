	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	arn := ackv1alpha1.AWSResourceName(databaseARN(ko))
	ko.Status.ACKResourceMetadata.ARN = &arn
