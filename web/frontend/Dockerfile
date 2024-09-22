# Build stage
FROM node:18 AS builder
WORKDIR /app

# Copy package.json and package-lock.json
COPY package*.json ./
RUN npm ci

# Copy the rest of the application code
COPY . .

# Build the Next.js app
RUN npm run build

# Production stage
FROM node:18-alpine
WORKDIR /app

# Copy built assets from builder stage
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/images ./images

# Expose port 3000
EXPOSE 3000

# Start the Next.js app
CMD ["npm", "start"]