	arn := ackv1alpha1.AWSResourceName(jobARN(ko))
	ko.Status.ACKResourceMetadata.ARN = &arn
	ko.Spec.Tags, err = rm.getTags(ctx, string(*ko.Status.ACKResourceMetadata.ARN))
