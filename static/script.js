// Обработка формы для отелей
document.getElementById('hotelForm')?.addEventListener('submit', function (e) {
    e.preventDefault();
    const destination = document.getElementById('hotel-destination').value;
    const checkIn = document.getElementById('hotel-checkin').value;
    const checkOut = document.getElementById('hotel-checkout').value;
    const guests = document.getElementById('hotel-guests').value;

    alert(`Searching hotels in ${destination} from ${checkIn} to ${checkOut} for ${guests} guests.`);
});

// Обработка формы для машин
document.getElementById('carForm')?.addEventListener('submit', function (e) {
    e.preventDefault();
    const pickup = document.getElementById('car-pickup').value;
    const pickupDate = document.getElementById('car-pickup-date').value;
    const dropoffDate = document.getElementById('car-dropoff-date').value;
    const carType = document.getElementById('car-type').value;

    alert(`Searching ${carType} cars for pickup at ${pickup} from ${pickupDate} to ${dropoffDate}.`);
});

// Обработка формы "Contact Us"
document.getElementById('contactForm')?.addEventListener('submit', async function (e) {
    e.preventDefault();

    const name = document.getElementById('name').value;
    const email = document.getElementById('email').value;
    const message = document.getElementById('message').value;

    if (!name || !email || !message) {
        alert("All fields are required.");
        return;
    }

    try {
        const response = await fetch('/contact', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ message: `${name}: ${email} - ${message}` }),
        });

        const result = await response.json();
        document.getElementById('response').innerText = JSON.stringify(result, null, 2);
    } catch (error) {
        console.error("Error submitting form:", error);
        document.getElementById('response').innerText = "Error submitting data.";
    }
});

// Динамическое отображение отелей
const hotels = [
    { name: "Luxury Inn", price: 200, rating: 4.5 },
    { name: "Economy Stay", price: 50, rating: 3.5 },
    { name: "Comfort Suites", price: 100, rating: 4.0 },
];

const hotelList = document.getElementById("hotelList");
const hotelSortSelect = document.getElementById("sortHotels");

const displayHotels = (hotels) => {
    if (hotelList) {
        hotelList.innerHTML = "";
        hotels.forEach(hotel => {
            hotelList.innerHTML += `
                <div class="hotel">
                    <h3>${hotel.name}</h3>
                    <p>Price: $${hotel.price} / night</p>
                    <p>Rating: ${hotel.rating} stars</p>
                </div>
            `;
        });
    }
};

hotelSortSelect?.addEventListener("change", () => {
    const criteria = hotelSortSelect.value;
    const sortedHotels = [...hotels];
    if (criteria === "price") sortedHotels.sort((a, b) => a.price - b.price);
    else if (criteria === "rating") sortedHotels.sort((a, b) => b.rating - a.rating);
    else if (criteria === "name") sortedHotels.sort((a, b) => a.name.localeCompare(b.name));
    displayHotels(sortedHotels);
});

displayHotels(hotels);


document.getElementById('loginForm').addEventListener('submit', function(e) {
    e.preventDefault();
    
    let email = document.getElementById('email').value;
    let password = document.getElementById('password').value;

    fetch('/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email: email, password: password })
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === "success") {
            window.location.href = '/profile'; // Перенаправление на страницу профиля
        } else {
            alert(data.message); // Показать сообщение об ошибке
        }
    })
    .catch(error => {
        console.error('Ошибка:', error);
    });
});


document.addEventListener("DOMContentLoaded", () => {
    const hotels = [
        { name: "Hotel California", price: 200, rating: 4.5 },
        { name: "Grand Budapest", price: 150, rating: 4.8 },
        { name: "The Plaza", price: 300, rating: 4.7 },
        { name: "Ritz Carlton", price: 350, rating: 4.9 },
    ];

    const hotelList = document.getElementById("hotelList");
    const filterInput = document.getElementById("filterHotels");
    const sortSelect = document.getElementById("sortHotels");
    const applyButton = document.getElementById("applyFilterSort");

    function renderHotels(filteredHotels) {
        hotelList.innerHTML = ""; // Clear the list
        filteredHotels.forEach((hotel) => {
            const hotelDiv = document.createElement("div");
            hotelDiv.classList.add("car");
            hotelDiv.innerHTML = `
                <h3>${hotel.name}</h3>
                <p>Price: $${hotel.price}</p>
                <p>Rating: ${hotel.rating}</p>
            `;
            hotelList.appendChild(hotelDiv);
        });
    }

    function applyFilterAndSort() {
        let filteredHotels = hotels;

        // Apply filter
        const filterText = filterInput.value.toLowerCase();
        if (filterText) {
            filteredHotels = filteredHotels.filter((hotel) =>
                hotel.name.toLowerCase().includes(filterText)
            );
        }

        // Apply sort
        const sortValue = sortSelect.value;
        if (sortValue === "price") {
            filteredHotels.sort((a, b) => a.price - b.price);
        } else if (sortValue === "rating") {
            filteredHotels.sort((a, b) => b.rating - a.rating);
        } else if (sortValue === "name") {
            filteredHotels.sort((a, b) => a.name.localeCompare(b.name));
        }

        renderHotels(filteredHotels);
    }

    applyButton.addEventListener("click", applyFilterAndSort);

    // Initial render
    renderHotels(hotels);
});


