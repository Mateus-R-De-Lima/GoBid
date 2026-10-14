APP_NAME=gobid
REGIO=N=us-east-1
VPC_ID=
SG_NAME=gobid-sg


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
 