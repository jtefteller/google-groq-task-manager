#!/bin/bash

mkdir $HOME/.google-groq-task-manager
touch $HOME/.google-groq-task-manager/token.json
touch $HOME/.google-groq-task-manager/.env
read -p "Enter your OAuth2 GOOGLE_REDIRECT_URL: " google_redirect_url 
read -p "Enter your OAuth2 GOOGLE_CLIENT_ID: " google_client_id 
read -p "Enter your OAuth2 GOOGLE_CLIENT_SECRET: " google_client_secret
read -p "Enter your GROQ_API_KEY: " groq_api_key
echo "GOOGLE_REDIRECT_URL=$google_redirect_url" >> $HOME/.google-groq-task-manager/.env
echo "GOOGLE_CLIENT_ID=$google_client_id" >> $HOME/.google-groq-task-manager/.env
echo "GOOGLE_CLIENT_SECRET=$google_client_secret" >> $HOME/.google-groq-task-manager/.env
echo "GROQ_API_KEY=$groq_api_key" >> $HOME/.google-groq-task-manager/.env
go mod tidy
go mod vendor
