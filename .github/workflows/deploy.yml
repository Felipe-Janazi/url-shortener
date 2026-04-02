name: Deploy

on:
  push:
    branches: [main]  # dispara apenas no merge para a main

jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      id-token: write  # necessário para autenticação OIDC com a AWS
      contents: read

    steps:
      - uses: actions/checkout@v4

      # Autenticação via OIDC: sem AWS_ACCESS_KEY_ID salva no repositório.
      # O GitHub prova sua identidade para a AWS via token temporário e assinado.
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1

      - name: Login to ECR
        id: ecr-login
        uses: aws-actions/amazon-ecr-login@v2

      - name: Build and push to ECR
        run: |
          IMAGE=${{ steps.ecr-login.outputs.registry }}/url-shortener:${{ github.sha }}
          docker build -t $IMAGE .
          docker push $IMAGE
          echo "IMAGE=$IMAGE" >> $GITHUB_ENV

      - name: Update ECS service
        run: |
          # Força o ECS a puxar a nova imagem e fazer rolling update.
          # O ECS cria novas tasks, valida o healthcheck e só derruba as antigas
          # quando as novas estiverem respondendo — zero downtime.
          aws ecs update-service \
            --cluster url-shortener-cluster \
            --service url-shortener-service \
            --force-new-deployment