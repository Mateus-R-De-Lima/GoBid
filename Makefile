APP_NAME=gobid
REGIO=N=us-east-1
VPC_ID=
SG_NAME=${APP_NAME}-sg
AWS_ACCOUNT_ID=$(shell aws sts get-caller-identity--query Accont --output text)
ECR_URL=$(AWS_ACCOUNT_ID).dkr.ecr.$(REGIO).amazonaws.com
REPO_URL=$(ECR_URL)/$(APP_NAME)
DB_NAME=${APP_NAME}-db

create-sg:
 @if ! aws ec2 describe-security-groups --filters Name=group-name,Values=$(SG_NAME)  --region $(REGIO) --query 'SecurityGroups[*].GroupId' --output text | grep -q . | grep -qE 'sg-'; then  \
    echo "Creating security group $(SG_NAME)..."; \
	aws ec2 create-security-group \
		--group-name $(SG_NAME) \
		--description "Security group for $(APP_NAME)" \
		--vpc-id $(VPC_ID) \
		--region $(REGIO); \
 else \ 
	echo "Security group $(SG_NAME) already exists."; \
	fi;
	SG_Id=$$(aws ec2 describe-security-groups --filters Name=group-name,Values=$(SG_NAME)  --region $(REGIO) --query 'SecurityGroups[*].GroupId' --output text); \
	echo  "Authorizing port 5432 on SG $$SG_ID..."; \
	aws ec2 authorize-security-group-ingress \
		--group-id $$SG_ID \
		--protocol tcp \
		--port 5432 \
		--cidr 0.0.0.0/0 \
		--region $(REGIO) || echo "Ingress rule already exists for port 5432 on SG $$SG_ID."; \

create-ecr:
	aws ecr describe-repositories --repository-name $(APP_NAME) --region $(REGIO) || \
	aws ecr create-repository --repository-name $(APP_NAME) --region $(REGIO)
 
build:
	docker build -t $(APP_NAME) -f Dockerfile.prod .

tag: build
	docker tag $(APP_NAME):latest $(REPO_URL):latest

push: tag
	aws acr get-login-password --region $(REGIO) | \
	docker login --username AWS --password-stdin $(ECR_URL)
	docker push $(REPO_URL):latest


create-db:
	@if ! aws rds describe-db-instances --db-instance-identifier ${DB_NAME} --region ${REGION} >/dev/null 2>&1; then \
		echo "Creating RDS instance ${DB_NAME}..."; \
		SG_ID=$$(aws ec2 describe-security-groups --filters Name=group-name,Values=${SG_NAME} --region ${REGION} --query 'SecurityGroups[0].GroupId' --output text); \
		aws rds create-db-instance \
			--db-instance-identifier ${DB_NAME} \
			--db-instance-class db.t3.micro \
			--engine postgres \
			--allocated-storage 20 \
			--master-username postgres \
			--master-user-password 123456 \
			--vpc-security-group-ids $$SG_ID \
			--publicly-accessible \
			--backup-retention-period 0 \
			--no-multi-az \
			--engine-version 11.22 \
			--port 5432 \
			--region ${REGION}; \
	else \
		echo "RDS instance ${DB_NAME} already exists."; \
	fi