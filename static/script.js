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


// Number of cars per page
const carsPerPage = 3;
document.addEventListener("DOMContentLoaded", function () {
    const cars = [
        { model: "Toyota Corolla", price: 50, rating: 4.5, category: "Sedan", brand: "Toyota" },
        { model: "Ford Explorer", price: 80, rating: 4.0, category: "SUV", brand: "Ford" },
        { model: "Tesla Model 3", price: 120, rating: 5.0, category: "Electric", brand: "Tesla" },
        { model: "Honda Civic", price: 40, rating: 4.2, category: "Sedan", brand: "Honda" },
        { model: "BMW XM", price: 200, rating: 5, category: "SUV", brand: "BMW" },
        { model: "Cadillac Escalade", price: 150, rating: 4.8, category: "SUV", brand: "Cadillac" },
        { model: "Rolls Royce Cullinan", price: 5000, rating: 5, category: "SUV", brand: "Rolls Royce" },
        { model: "Mercedes G63", price: 300, rating: 4.9, category: "SUV", brand: "Mercedes" },
        { model: "Mercedes GLE53", price: 150, rating: 4.5, category: "SUV", brand: "Mercedes" },
        { model: "GMC SLT", price: 100, rating: 4.0, category: "SUV", brand: "GMC" },
        { model: "Porsche Macan", price: 300, rating: 4.7, category: "SUV", brand: "Porsche" },
        { model: "Nissan Patrol", price: 100, rating: 4.2, category: "SUV", brand: "Nissan" },
        { model: "BMW M4 Competition", price: 200, rating: 4.8, category: "Sedan", brand: "BMW" },
        { model: "Audi RS3", price: 220, rating: 4.6, category: "Sedan", brand: "Audi" },
        { model: "Audi RS5", price: 270, rating: 4.7, category: "Sedan", brand: "Audi" },
        { model: "Audi S8", price: 300, rating: 4.9, category: "Sedan", brand: "Audi" },
        { model: "BMW 730LI", price: 290, rating: 4.6, category: "Sedan", brand: "BMW" },
        { model: "Mercedes EQE 350", price: 120, rating: 4.5, category: "Electric", brand: "Mercedes" },
        { model: "Tesla Model 3", price: 120, rating: 5.0, category: "Electric", brand: "Tesla" },
        { model: "Porsche 718", price: 4718, rating: 4.9, category: "Sports", brand: "Porsche" },
        { model: "Porsche 911 Turbo S", price: 9000, rating: 5.0, category: "Sports", brand: "Porsche" },
        { model: "Ferrari F8 Tributo", price: 9999, rating: 5.0, category: "Sports", brand: "Ferrari" },
        { model: "Audi R8", price: 2000, rating: 4.8, category: "Sports", brand: "Audi" },
        { model: "Audi RS6", price: 300, rating: 4.7, category: "Sports", brand: "Audi" },
        { model: "Mercedes V250", price: 2500, rating: 4.6, category: "Van", brand: "Mercedes" }
    ];

    const carList = document.getElementById("carList");
    const carCategoryFilter = document.getElementById("carCategory");
    const carBrandFilter = document.getElementById("carBrand");
    const carSortFilter = document.getElementById("carSort");
    const paginationControls = document.getElementById("paginationControls");

    let currentPage = 1;
    const carsPerPage = 3;

    // Function to display cars
    function displayCars(filteredCars, page = 1) {
        carList.innerHTML = "";
        const startIndex = (page - 1) * carsPerPage;
        const endIndex = startIndex + carsPerPage;
        const paginatedCars = filteredCars.slice(startIndex, endIndex);

        paginatedCars.forEach(car => {
            const carCard = document.createElement("div");
            carCard.classList.add("col-md-4", "car-card");
            carCard.innerHTML = `
                <h3>${car.model}</h3>
                <p>Price: $${car.price} / day</p>
                <p>Rating: ${car.rating} stars</p>
            `;
            carList.appendChild(carCard);
        });

        renderPagination(filteredCars.length, page);
    }

    // Function to render pagination controls
    // Function to render pagination controls
function renderPagination(totalCars, currentPage) {
    const totalPages = Math.ceil(totalCars / carsPerPage);
    paginationControls.innerHTML = "";

    if (totalPages > 1) {
        for (let i = 1; i <= totalPages; i++) {
            const button = document.createElement("button");
            button.classList.add("btn", "mx-1");
            button.textContent = i;
            button.onclick = () => {
                currentPage = i;
                filterCars(currentPage);
            };
            if (i === currentPage) {
                button.classList.add("active");
            }
            paginationControls.appendChild(button);
        }
    }
}


    // Filter and sort function
    function filterCars(page = 1) {
        const selectedCategory = carCategoryFilter.value;
        const selectedBrand = carBrandFilter.value;

        let filteredCars = cars.filter(car => {
            const categoryMatch = selectedCategory === "All Categories" || car.category === selectedCategory;
            const brandMatch = selectedBrand === "All Brands" || car.brand === selectedBrand;
            return categoryMatch && brandMatch;
        });

        const selectedSort = carSortFilter.value;
        if (selectedSort === "rating") {
            filteredCars = filteredCars.sort((a, b) => b.rating - a.rating);
        } else if (selectedSort === "price") {
            filteredCars = filteredCars.sort((a, b) => a.price - b.price);
        }

        displayCars(filteredCars, page);
    }

    // Event listeners for filters
    carCategoryFilter.addEventListener("change", () => filterCars(currentPage));
    carBrandFilter.addEventListener("change", () => filterCars(currentPage));
    carSortFilter.addEventListener("change", () => filterCars(currentPage));

    // Initial display of cars
    filterCars();
});




// Динамическое отображение автомобилей


const carList = document.getElementById("carList");
const carSortSelect = document.getElementById("sortCars");

const displayCars = (cars) => {
    if (carList) {
        carList.innerHTML = "";
        cars.forEach(car => {
            carList.innerHTML += `
                <div class="car">
                    <h3>${car.model}</h3>
                    <p>Price: $${car.price} / day</p>
                    <p>Rating: ${car.rating} stars</p>
                </div>
            `;
        });
    }
};

carSortSelect?.addEventListener("change", () => {
    const criteria = carSortSelect.value;
    const sortedCars = [...cars];
    if (criteria === "price") sortedCars.sort((a, b) => a.price - b.price);
    else if (criteria === "rating") sortedCars.sort((a, b) => b.rating - a.rating);
    else if (criteria === "model") sortedCars.sort((a, b) => a.model.localeCompare(b.model));
    displayCars(sortedCars);
});

displayCars(cars);

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


