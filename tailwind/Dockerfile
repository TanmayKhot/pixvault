FROM node:latest

RUN mkdir -p /src /dst
WORKDIR /src

# Copy package.json and install dependencies
COPY package.json ./
RUN npm install

# Copy only the necessary files
COPY tailwind.config.js .
COPY styles.css .

# Run tailwindcss using the full path to the binary
CMD ["./node_modules/.bin/tailwindcss", "-c", "tailwind.config.js", "-i", "styles.css", "-o", "/dst/styles.css", "--watch", "--poll"]