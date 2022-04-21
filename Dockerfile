FROM golang:1.17

ARG cot_text_encryption=false
ARG cot_public_key_file=
ARG cot_private_key_file=
ARG cot_passphrase=
ARG cot_cn_public_key_dir=/cot/cn_secrets
ARG cot_sig_verification=
ARG cot_base64_encoding=
ARG cot_conf_dir=/cot/config

ENV COT_TEXT_ENCRYPTION=${cot_text_encryption}
ENV COT_PUBLIC_KEY_FILE=${cot_public_key_file}
ENV COT_PRIVATE_KEY_FILE=${cot_private_key_file}
ENV COT_PASSPHRASE=${cot_passphrase}
ENV COT_CN_PUBLIC_KEY_DIR=${cot_cn_public_key_dir}
ENV COT_SIG_VERIFICATION=${cot_sig_verification}
ENV COT_BASE64_ENCODING=${cot_base64_encoding}
ENV COT_CONF_DIR=${cot_conf_dir}

WORKDIR /go/src/app
COPY . .

RUN mkdir -p /cot/config /cot/cn_secrets /cot/secrets

RUN go get -d -v ./...
RUN go install -v ./...
RUN go build

VOLUME /cot/config
VOLUME /cot/cn_secrets
VOLUME /cot/secrets

CMD ["cot", "-logtostderr=true"]