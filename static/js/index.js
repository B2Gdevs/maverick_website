let fileBrow = document.getElementById("customFile");
let formSubmitButton = document.getElementById("submitButton");

fileBrow.addEventListener("change", () => {
    document.getElementById("customFileLabel").innerText = fileBrow.files[0].name;
})

formSubmitButton.addEventListener("click", () => {
    let customFile = document.getElementById("customFile");
    let postTitle = document.getElementById("post-title");
    let description = document.getElementById("description");
    let postError = document.getElementById("postError");

    if (customFile.files[0].size > 16 * 1024 * 1024){
        postError.innerText = `Files must be less than 16MB.  Your file is ${Math.floor(customFile.files[0].size/1024/1024)}.`
        postError.className = "d-block"
    } else{
        postError.innerText = `Upload Error: Need to have an image or video to upload a post!`
        postError.className = "d-none"
    }

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
            postError.className = "d-none";
        } else{
            postError.className = "d-block";
        }
        
    })

    
})
