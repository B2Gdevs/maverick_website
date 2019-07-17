let fileBrow = document.getElementById("customFile");
let formSubmitButton = document.getElementById("submitButton");

fileBrow.addEventListener("change", () => {
    document.getElementById("customFileLabel").innerText = fileBrow.files[0].name;
})

formSubmitButton.addEventListener("click", () => {
    let customFile = document.getElementById("customFile");
    let postTitle = document.getElementById("post-title");
    let description = document.getElementById("description");

    let data = new FormData();
    data.append("CustomFile", customFile.files[0]);
    data.append("Title", postTitle.value);
    data.append("Description", description.value);

    fetch("/Media", {
        method: "POST",
        body: data
    }).then( (response) =>{
        console.log(response.text());
        if (response.ok){
            location.reload(false);
            document.getElementById("postError").className = "d-none"
        }
        document.getElementById("postError").className = "d-block";
    })

    
})
