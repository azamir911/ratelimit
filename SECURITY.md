# Security policy

Please do not open a public issue for a vulnerability that could put users at risk. Report security concerns privately through GitHub's private vulnerability reporting feature when it is available for this repository.

The limiter is process-local and should not be treated as an authentication or authorization control. Callers are responsible for deriving trustworthy keys and for choosing fail-open or fail-closed behavior when the limiter reports capacity or lifecycle errors.
