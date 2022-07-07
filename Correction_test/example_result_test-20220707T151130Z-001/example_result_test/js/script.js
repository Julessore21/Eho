const main = document.getElementById("main");

function parseJson(data) {
	return JSON.parse(data);
}

function showPictures(pictures) {
	parseJson(data);

	main.innerHTML = "";

	pictures.data.forEach((picture) => {
		console.log(picture);
	});

	pictures.data.forEach((picture) => {
		
		let { img_path, title, vote_average, explanation } = picture;

		let pictureEl = document.createElement("div");
		pictureEl.classList.add("picture");

		pictureEl.innerHTML = `
            <img
                src="${img_path}"
                alt="${title}"
            />
            <div class="picture-info">
                <h3>${title}</h3>
                <span class="${getClassByRate(
			vote_average
		)}">${vote_average}</span>
            </div>
            <div class="explanation">
                <h3>Explication:</h3>
                ${explanation}
            </div>
        `;

		main.appendChild(pictureEl);
	});
}

function getClassByRate(vote) {
	if (vote >= 8) {
		return "green";
	} else if (vote >= 5) {
		return "orange";
	} else {
		return "red";
	}
}

function init() {
	let obj = parseJson(data, true);
	showPictures(obj);
}

init();
