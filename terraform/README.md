# URL Shortener AWS Deployment with Terraform

This directory contains Terraform configurations to deploy the URL Shortener application to AWS using free tier resources.

## Architecture

The deployment includes:
- **VPC**: Custom VPC with public and private subnets across 2 availability zones
- **EC2**: Auto Scaling Group with t2.micro instances (free tier)
- **RDS**: PostgreSQL database (db.t3.micro, free tier)
- **ElastiCache**: Redis cluster (cache.t2.micro, free tier)
- **ALB**: Application Load Balancer for high availability
- **Security Groups**: Proper security configurations

## Prerequisites

1. **AWS Account**: Ensure you have an AWS account with free tier available
2. **AWS CLI**: Install and configure AWS CLI with your credentials
3. **Terraform**: Install Terraform (>= 1.0)
4. **Key Pair**: Create an EC2 key pair in your AWS region

## Quick Start

1. **Configure AWS Credentials**:
   ```bash
   aws configure
   ```

2. **Copy and Edit Variables**:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your values
   ```

3. **Initialize Terraform**:
   ```bash
   terraform init
   ```

4. **Plan Deployment**:
   ```bash
   terraform plan
   ```

5. **Deploy Infrastructure**:
   ```bash
   terraform apply
   ```

## Configuration

### Required Variables

Edit `terraform.tfvars` with these required values:

- `key_pair_name`: Your AWS EC2 key pair name
- `db_password`: Secure password for PostgreSQL database

### Optional Configurations

- `aws_region`: AWS region (default: us-east-1)
- `instance_type`: EC2 instance type (default: t2.micro)
- `app_port`: Application port (default: 8080)
- `redis_auth_token`: Redis authentication token (leave empty for no auth)
- `allowed_cidr_blocks`: IP ranges allowed to access the app (default: 0.0.0.0/0)

## Deployment Process

After running `terraform apply`, the infrastructure will be created. However, you'll need to deploy your application code separately.

### Application Deployment Options

1. **Manual Deployment via SSH**:
   - SSH into the EC2 instance
   - Clone your repository
   - Build and run your Docker containers

2. **GitHub Actions CI/CD** (Recommended):
   - Use the existing CI/CD pipeline
   - Configure AWS credentials as GitHub secrets
   - Deploy to EC2 instances automatically

3. **AWS CodeDeploy**:
   - Set up CodeDeploy for automated deployments

## Free Tier Resources

This configuration uses the following free tier eligible resources:

- **EC2**: t2.micro instances (750 hours/month)
- **RDS**: db.t3.micro PostgreSQL (750 hours/month, 20 GB storage)
- **ElastiCache**: cache.t2.micro Redis (750 hours/month)
- **ALB**: Application Load Balancer (750 hours/month)

## Security Considerations

⚠️ **Important Security Notes**:

1. **Database Password**: Use a strong password and consider using AWS Secrets Manager
2. **Access Control**: Restrict `allowed_cidr_blocks` to your IP ranges in production
3. **Key Pair**: Keep your EC2 key pair secure
4. **Redis Auth**: Consider enabling Redis authentication for production use

## Monitoring and Logs

- CloudWatch agent is installed on EC2 instances for monitoring
- Application logs are stored in `/var/log/url-shortener`
- Use AWS CloudWatch for monitoring and alerting

## SSL/HTTPS (Optional)

To enable HTTPS:

1. Register a domain name
2. Request an SSL certificate via AWS Certificate Manager
3. Uncomment the HTTPS listener configuration in `load_balancer.tf`
4. Update DNS records to point to the load balancer

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

⚠️ **Warning**: This will delete all data in the RDS database!

## Cost Optimization

- All resources are configured for free tier eligibility
- Auto Scaling Group keeps minimum instances to 1
- RDS backup retention is set to 7 days (minimum)
- No expensive features like Multi-AZ or encryption are enabled

## Troubleshooting

### Common Issues

1. **Key Pair Not Found**: Ensure your key pair exists in the specified AWS region
2. **Free Tier Limits**: Check your AWS free tier usage in the billing console
3. **Security Group Issues**: Verify security group rules allow necessary traffic

### Useful Commands

```bash
# Check Terraform state
terraform show

# View outputs
terraform output

# SSH to instances (replace with actual IP)
ssh -i your-key.pem ec2-user@instance-ip

# View application logs
docker logs url-shortener_app_1
```

## Support

For issues related to:
- **Terraform**: Check the official Terraform documentation
- **AWS Resources**: Consult AWS documentation
- **Application**: Check the main README.md in the project root
