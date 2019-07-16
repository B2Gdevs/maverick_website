var fileBrow = document.getElementById("customFile")

fileBrow.addEventListener("change", () => {
    document.getElementById("customFileLabel").innerText = fileBrow.files[0].name
})
