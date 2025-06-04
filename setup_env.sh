docker build -t rocky8-localrepo .
docker run -it -v $(pwd):/app rocky8-localrepo