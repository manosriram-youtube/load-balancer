FROM node:12.9

WORKDIR /app

RUN git clone https://github.com/manosriram/express-template.git
WORKDIR /app/express-template

COPY . .

RUN npm install
RUN apt-get update && apt-get install -y vim

COPY . .
CMD ["node", "index.js"]
