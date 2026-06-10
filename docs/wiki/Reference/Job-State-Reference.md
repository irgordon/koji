# Job State Reference

[Home](../Home.md) | Related: [Job Lifecycle](../Architecture/Job-Lifecycle.md), [Jobs Page](../User-Guide/Jobs-Page.md)

Semi-generated from `internal/jobs`.

| State | Meaning | Valid Next States |
| --- | --- | --- |
| `queued` | Service-control intent is waiting for human decision. | `approved`, `rejected` |
| `approved` | A human approved the job. | `running` |
| `rejected` | A human rejected the job. | Terminal |
| `running` | Worker claimed the approved job. | `completed`, `failed`, `not_implemented` |
| `completed` | Agent path completed successfully. | Terminal |
| `failed` | Worker or agent path failed safely. | Terminal |
| `not_implemented` | Agent reported unsupported behavior. | Terminal |

## Limits

Default job list limit is 50. Maximum list limit is 100.

## Recovery

Terminal failed jobs are not retried in place. Fix the cause and create a new service-control request.
