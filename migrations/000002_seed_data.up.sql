-- Seed data for development and testing

-- Insert sample training modules

-- S3 Basics
INSERT INTO training_modules (name, title, description, category, difficulty, estimated_minutes, content) VALUES
('s3-basics', 'S3 Basics', 'Learn the fundamentals of Amazon S3 object storage', 's3', 'beginner', 15,
'{"sections": [
  {
    "id": "intro",
    "title": "Introduction to Amazon S3",
    "type": "text",
    "content": "Amazon S3 (Simple Storage Service) is a highly scalable, durable, and secure object storage service provided by AWS. Unlike traditional file systems, S3 stores data as objects within containers called buckets.\n\n**Common Use Cases:**\n- Backup and archival of research data\n- Data lakes for analytics and machine learning\n- Static website hosting\n- Content distribution and media storage\n- Disaster recovery\n\n**Why S3 for Research:**\nS3 provides 99.999999999% (11 9s) durability, meaning your research data is extremely safe. It scales automatically from gigabytes to petabytes without any infrastructure management. You only pay for what you use, making it cost-effective for projects of any size."
  },
  {
    "id": "concepts",
    "title": "Key Concepts",
    "type": "text",
    "content": "**Buckets:**\nBuckets are containers for objects stored in S3. Each bucket must have a globally unique name across all AWS accounts. Bucket names must be 3-63 characters, contain only lowercase letters, numbers, hyphens, and periods, and cannot be formatted like an IP address.\n\n**Objects:**\nObjects are the fundamental entities stored in S3. An object consists of data (the file itself), metadata (information about the file), and a unique key. Individual objects can range from 0 bytes to 5 TB in size.\n\n**Keys:**\nA key is the unique identifier for an object within a bucket. The key is the full path to the object, such as ''data/2025/experiment-results.csv''. Keys can include forward slashes to create a logical folder structure.\n\n**Regions:**\nWhen you create a bucket, you choose an AWS Region where S3 will store the bucket and its objects. Choose a region close to your users or compute resources to minimize latency and costs. Data never leaves your chosen region unless you explicitly transfer it."
  },
  {
    "id": "security",
    "title": "Security & Encryption",
    "type": "text",
    "content": "**Bucket Policies vs IAM Policies:**\nBucket policies are attached to S3 buckets and define who can access the bucket and what actions they can perform. IAM policies are attached to users, groups, or roles and define what AWS resources they can access. For research projects, you typically use both: IAM policies to grant your team members access, and bucket policies to enforce additional restrictions.\n\n**Server-Side Encryption:**\nS3 offers multiple encryption options to protect your data at rest:\n- **AES256 (SSE-S3):** Amazon manages encryption keys automatically. This is the simplest and most common option.\n- **AWS KMS (SSE-KMS):** You control encryption keys through AWS Key Management Service, providing additional audit trails and access controls.\n\n**Versioning:**\nVersioning keeps multiple versions of an object in the same bucket. When enabled, S3 preserves, retrieves, and restores every version of every object. This protects against accidental deletion and overwrites. Once enabled, versioning cannot be fully disabled—only suspended.\n\n**Block Public Access:**\nBy default, S3 buckets and objects are private. AWS provides Block Public Access settings at both the account and bucket level to prevent accidental public exposure. For research data, you should typically keep these settings enabled unless you specifically need to share data publicly."
  },
  {
    "id": "costs",
    "title": "Cost Considerations",
    "type": "text",
    "content": "**Storage Classes:**\nS3 offers multiple storage classes optimized for different access patterns:\n- **S3 Standard:** For frequently accessed data. Costs ~$0.023/GB/month.\n- **S3 Infrequent Access (IA):** For data accessed less than once per month. Costs ~$0.0125/GB/month.\n- **S3 Glacier:** For long-term archival with retrieval times from minutes to hours. Costs ~$0.004/GB/month.\n- **S3 Glacier Deep Archive:** For data accessed once or twice per year. Costs ~$0.00099/GB/month.\n\n**Data Transfer Costs:**\n- Data transferred INTO S3 is free\n- Data transferred OUT to the internet costs ~$0.09/GB (first 10 TB/month)\n- Data transferred between S3 and EC2 in the same region is free\n\n**Request Costs:**\nS3 charges for API requests:\n- PUT, COPY, POST, LIST requests: ~$0.005 per 1,000 requests\n- GET, SELECT requests: ~$0.0004 per 1,000 requests\n\n**Cost Optimization Tips:**\n- Use lifecycle policies to automatically transition older data to cheaper storage classes\n- Delete incomplete multipart uploads\n- Enable S3 Intelligent-Tiering for unpredictable access patterns\n- Use S3 Storage Class Analysis to understand your access patterns"
  },
  {
    "id": "best-practices",
    "title": "Best Practices",
    "type": "text",
    "content": "**1. Use Descriptive Bucket Names:**\nChoose bucket names that clearly indicate the project and purpose, such as ''mylab-genomics-raw-data'' or ''research-project-alpha-results''. Remember that bucket names are globally unique and cannot be changed after creation.\n\n**2. Enable Versioning for Important Data:**\nFor any research data that cannot be easily recreated, enable versioning. This protects against accidental deletion and provides an audit trail of changes.\n\n**3. Use Encryption by Default:**\nAlways enable server-side encryption for research data. Start with AES256 (SSE-S3) for simplicity, or use AWS KMS if you need additional access controls and audit trails.\n\n**4. Implement Lifecycle Policies:**\nSet up lifecycle policies to automatically transition older data to cheaper storage classes. For example, move data to Glacier after 90 days if it is rarely accessed.\n\n**5. Tag Resources for Cost Tracking:**\nUse tags to organize and track costs by project, department, or principal investigator. Tags like ''Project=Alpha'', ''PI=DrSmith'', ''Grant=NSF-12345'' make it easy to allocate costs and generate reports.\n\n**6. Use Folders for Organization:**\nWhile S3 does not have true folders, you can use key prefixes (paths) to organize objects logically. For example: ''raw-data/2025/01/'', ''processed-data/experiments/'', ''results/publication-ready/''.\n\n**7. Set Up Bucket Logging:**\nEnable S3 server access logging to track all requests made to your bucket. This is valuable for security audits and understanding access patterns.\n\n**8. Review Permissions Regularly:**\nPeriodically review bucket policies and IAM permissions to ensure only authorized users have access. Remove access for users who no longer need it."
  },
  {
    "id": "quiz",
    "title": "Knowledge Check",
    "type": "quiz",
    "questions": [
      {
        "id": "q1",
        "question": "What is the maximum size of a single object that can be stored in S3?",
        "options": [
          "5 GB",
          "5 TB",
          "50 TB",
          "No limit"
        ],
        "correctAnswer": 1,
        "explanation": "The maximum object size in S3 is 5 TB. Objects larger than 5 GB must be uploaded using multipart upload."
      },
      {
        "id": "q2",
        "question": "Which encryption type requires you to manage keys through AWS Key Management Service (KMS)?",
        "options": [
          "AES256 (SSE-S3)",
          "AWS KMS (SSE-KMS)",
          "Both",
          "Neither"
        ],
        "correctAnswer": 1,
        "explanation": "SSE-KMS uses AWS Key Management Service for key management, giving you control over encryption keys and providing detailed audit trails."
      },
      {
        "id": "q3",
        "question": "What happens to old versions of an object when versioning is enabled?",
        "options": [
          "Old versions are automatically deleted",
          "Old versions are kept and can be retrieved",
          "Objects cannot be overwritten",
          "Only the last 10 versions are kept"
        ],
        "correctAnswer": 1,
        "explanation": "When versioning is enabled, S3 keeps all versions of an object. Old versions remain accessible and can be retrieved, restored, or permanently deleted if needed."
      },
      {
        "id": "q4",
        "question": "Which S3 storage class is most cost-effective for long-term archival data that is accessed once or twice per year?",
        "options": [
          "S3 Standard",
          "S3 Infrequent Access",
          "S3 Glacier",
          "S3 Glacier Deep Archive"
        ],
        "correctAnswer": 3,
        "explanation": "S3 Glacier Deep Archive is the lowest-cost storage class, designed for data that is accessed once or twice per year with retrieval times of 12-48 hours."
      },
      {
        "id": "q5",
        "question": "Can you change a bucket name after the bucket has been created?",
        "options": [
          "Yes, at any time",
          "Yes, but only within 24 hours of creation",
          "No, bucket names cannot be changed",
          "Yes, but only if the bucket is empty"
        ],
        "correctAnswer": 2,
        "explanation": "Bucket names cannot be changed after creation. If you need a different name, you must create a new bucket and copy the objects to it."
      }
    ]
  }
]}'::jsonb);

-- S3 Security
INSERT INTO training_modules (name, title, description, category, difficulty, estimated_minutes, content, prerequisites) VALUES
('s3-security', 'S3 Security Best Practices', 'Learn how to secure your S3 buckets and data', 's3', 'intermediate', 20,
'{"sections": [
  {"title": "Bucket Policies", "content": "Control access to your buckets using bucket policies."},
  {"title": "Encryption", "content": "Encrypt data at rest using SSE-S3, SSE-KMS, or SSE-C."},
  {"title": "Access Logging", "content": "Enable server access logging to track requests."},
  {"title": "Versioning", "content": "Enable versioning to protect against accidental deletion."}
]}'::jsonb,
ARRAY['s3-basics']);

-- IAM Basics
INSERT INTO training_modules (name, title, description, category, difficulty, estimated_minutes, content) VALUES
('iam-basics', 'IAM Fundamentals', 'Understand AWS Identity and Access Management', 'iam', 'beginner', 20,
'{"sections": [
  {"title": "What is IAM?", "content": "IAM enables you to manage access to AWS services and resources securely."},
  {"title": "Users and Groups", "content": "Create IAM users and organize them into groups."},
  {"title": "Roles", "content": "Use IAM roles to grant temporary access to AWS resources."},
  {"title": "Policies", "content": "Define permissions using JSON policy documents."}
]}'::jsonb);

-- EC2 Basics
INSERT INTO training_modules (name, title, description, category, difficulty, estimated_minutes, content) VALUES
('ec2-basics', 'EC2 Fundamentals', 'Learn about Amazon EC2 virtual servers', 'ec2', 'beginner', 25,
'{"sections": [
  {"title": "What is EC2?", "content": "Amazon EC2 provides scalable computing capacity in the cloud."},
  {"title": "Instance Types", "content": "Choose the right instance type for your workload."},
  {"title": "Security Groups", "content": "Control inbound and outbound traffic with security groups."},
  {"title": "Key Pairs", "content": "Use key pairs for secure SSH access to instances."}
]}'::jsonb);

-- Insert sample policies

-- Training gate for S3
INSERT INTO policies (name, description, policy_type, rules, applies_to) VALUES
('s3-training-gate', 'Require S3 basics training before creating buckets', 'training_gate',
'{"required_modules": ["s3-basics"], "actions": ["s3:CreateBucket"]}'::jsonb,
ARRAY['researcher']);

-- Resource limit
INSERT INTO policies (name, description, policy_type, rules, applies_to) VALUES
('s3-bucket-limit', 'Limit number of S3 buckets per user', 'resource_limit',
'{"resource_type": "s3:bucket", "max_count": 10}'::jsonb,
ARRAY['researcher']);

-- EC2 instance limit
INSERT INTO policies (name, description, policy_type, rules, applies_to) VALUES
('ec2-instance-limit', 'Limit number of EC2 instances per user', 'resource_limit',
'{"resource_type": "ec2:instance", "max_count": 5, "max_total_vcpus": 16}'::jsonb,
ARRAY['researcher']);
