package mock

var (
	OkIndex = `	<!DOCTYPE html>
				<head>
					<title>Title</title>
				</head>
				<body>
					<a href="/1/">1</a>
				</body>
				</html>`

	OkPage1 = `	<!DOCTYPE html>
				<head>
					<title>Title</title>
				</head>
				<body>
					<a href="/2/">2</a>
				</body>
				</html>`

	OkPage2 = `	<!DOCTYPE html>
				<head>
					<title>Title</title>
				</head>
				<body>
					<a href="/">index</a>
					<a href="/1/">1</a>
				</body>
				</html>`

	OkPage2Redirected = `	<!DOCTYPE html>
							<head>
								<title>Title</title>
							</head>
							<body>
								<a href="/">index</a>
							</body>
							</html>`
)
