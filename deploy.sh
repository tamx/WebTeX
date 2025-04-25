CLOUDSDK_CORE_PROJECT=tamcloud \
    gcloud builds submit --config cloudbuild.yaml \
    --substitutions _SERVICE_NAME=webtex
