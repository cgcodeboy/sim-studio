FROM node:22-alpine
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
RUN npm run build
ENV API_ORIGIN=http://api:8080
EXPOSE 8080
CMD ["node", "server.js"]
