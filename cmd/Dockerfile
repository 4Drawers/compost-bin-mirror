FROM ubuntu

WORKDIR /usr/src

COPY ./build/cmd ./
RUN chmod +x ./cmd

CMD [ "/usr/src/cmd" ]

EXPOSE 17890