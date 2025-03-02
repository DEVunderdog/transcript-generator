# How GCP cloud services are leverage
![image](https://github.com/user-attachments/assets/bccc90f7-47d6-4091-8a51-6232872be041)

- I wanted to independent services as backend as a service and transcript-service as a service.
- Clients could directly interact with backend which is Public API to generate transcript. 
- And GCP Pub/Sub is how we try to maintain communication between two decoupled services.
- Now We have ran a Cloud SQL instance of Postgres version 16 on GCP, which would interact with backend service which is hosted on cloud run over private network.
- The database isn't exposed publicly, hence on VPC but our backend service which is accessible over internet needed a  way to communicate internally to the database, hence we had a VPC connector on that Cloud Run,
  so that via private routing of traffic they could maintain connection.
- Another fascinated thing about transcript_service which is that its been isolated, not hosted over internet and has VPC connector for its communication internally.
- And we explicitly connect Pub/Sub message listening capabilities to it. 
