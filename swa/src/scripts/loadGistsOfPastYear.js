async function loadGistsOfPastYear() {
  const headers = {
    "Accept": 'application/vnd.github+json',
    "X-GitHub-Api-Version": "2022-11-28",
  }

  if (process.env.GITHUB_TOKEN) {
    headers['Authorization'] = `Bearer ${process.env.GITHUB_TOKEN}`;
  }

  // Get ISO 8601 formatted date string for the last year
  const lastYear = new Date();
  lastYear.setFullYear(lastYear.getFullYear() - 1);
  const lastYearString = lastYear.toISOString();

  const response = await fetch(`https://api.github.com/users/clowa/gists?since=${lastYearString}`, {
    headers: headers,
  });
  if (!response.ok) {
    throw new Error(`Network response was not ok ${response.statusText}`);
  }
  const gists = await response.json();
  // Print properties description and html_url for gists to console
  gists.forEach(gist => {
    console.log(gist.description, gist.html_url);
  });

  return gists;
}

export default loadGistsOfPastYear;
