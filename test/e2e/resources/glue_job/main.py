import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job

# Create Glue context
glueContext = GlueContext(SparkContext.getOrCreate())

# Create job
job = Job(glueContext)
job.init("minimal_test_job")

print("Hello from Glue!")

job.commit()