import { AzureFunction, Context, HttpRequest } from "@azure/functions"

interface IQoute {
  id: string
  content: string
  author: string
  authorSlug: string
  length: number
  tags: string[]
  creationdate: Date
}

const httpTrigger: AzureFunction = async function (
  context: Context,
  req: HttpRequest
): Promise<void> {
  context.log("HTTP trigger function processed a request.")
  console.log("Received request: ", req.headers.host)

  console.log("Check presence of environment variable API_URL")
  const api_url = process.env.API_URL
  if (!api_url) {
    console.error("Environment variable API_URL is not set.")
    context.res = {
      status: 500,
      body: "Internal Server Error",
    }
    return
  }

  console.log("Check presence of environment variable API_KEY")
  const api_key = process.env.API_KEY
  if (!api_key) {
    console.error("Environment variable API_KEY is not set.")
    context.res = {
      status: 500,
      body: "Internal Server Error",
    }
    return
  }

  // Creating the API request
  const headers = new Headers({
    "Content-Type": "application/json",
    Accept: "application/json",
    "Ocp-Apim-Subscription-Key": api_key,
  })

  try {
    console.log("Fetching quote from API.")
    const response = await fetch(api_url, {
      method: "GET",
      headers: headers,
    })

    if (!response.headers.get("Content-Type")?.includes("application/json")) {
      console.error("API did not return JSON response.")
      context.res = {
        status: 500,
        body: "Internal Server Error",
      }
      return
    }

    const respJson = response.json()
    const quote = (await respJson) as IQoute

    context.res = {
      Headers: {
        "Content-Type": "plain/text",
        "Allow-Control-Allow-Origin": "*",
      },
      status: 200 /* Defaults to 200 */,
      body: {
        content: quote.content,
        author: quote.author,
        creationdate: quote.creationdate,
      },
    }
  } catch (err) {
    console.error("Failed to call API: ", err)

    context.res = {
      status: 500,
      body: "Internal Server Error",
    }
  }
}

export default httpTrigger
