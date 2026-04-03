import jakarta.validation.Valid;
import jakarta.validation.constraints.Min;
import lombok.Builder;
import lombok.Data;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.data.domain.*;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.bind.annotation.*;

import java.math.BigDecimal;
import java.util.List;
import java.util.Optional;

// --- DTOs ---

@Data
class ProductRequest {
    private String name;
    private String description;
    private BigDecimal price;
}

@Data
@Builder
class ProductResponse {
    private Long id;
    private String name;
    private String description;
    private BigDecimal price;
}

@Data
class ProductSearchCriteria {
    private String name;
    private BigDecimal minPrice;
    private BigDecimal maxPrice;
}

@Data
@Builder
class PagedResponse<T> {
    private List<T> content;
    private int page;
    private int size;
    private long totalElements;
}

// --- Entity ---

@jakarta.persistence.Entity
@jakarta.persistence.Table(name = "products")
@Data
class Product {
    @jakarta.persistence.Id
    @jakarta.persistence.GeneratedValue(strategy = jakarta.persistence.GenerationType.IDENTITY)
    private Long id;
    private String name;
    private String description;
    private BigDecimal price;
}

// --- Repository ---

interface ProductRepository extends JpaRepository<Product, Long>, JpaSpecificationExecutor<Product> {
    List<Product> findByNameContainingIgnoreCase(String name);
}

// --- Service ---

@Service
class ProductService {
    private final ProductRepository repository;

    public ProductService(ProductRepository repository) {
        this.repository = repository;
    }

    @Cacheable(value = "products", key = "#id")
    public ProductResponse getProduct(Long id) {
        Product p = repository.findById(id).orElseThrow(() -> new RuntimeException("Product not found"));
        return mapToResponse(p);
    }

    public PagedResponse<ProductResponse> search(ProductSearchCriteria criteria, int page, int size, String sort) {
        Pageable pageable = PageRequest.of(page, size, Sort.by(sort));
        Specification<Product> spec = (root, query, cb) -> {
            var predicates = cb.conjunction();
            if (criteria.getName() != null) predicates = cb.and(predicates, cb.like(root.get("name"), "%" + criteria.getName() + "%"));
            if (criteria.getMinPrice() != null) predicates = cb.and(predicates, cb.greaterThanOrEqualTo(root.get("price"), criteria.getMinPrice()));
            return predicates;
        };
        Page<Product> paged = repository.findAll(spec, pageable);
        return PagedResponse.<ProductResponse>builder()
                .content(paged.getContent().stream().map(this::mapToResponse).toList())
                .page(page).size(size).totalElements(paged.getTotalElements()).build();
    }

    @CacheEvict(value = "products", allEntries = true)
    public ProductResponse create(ProductRequest request) {
        Product p = new Product();
        p.setName(request.getName());
        p.setPrice(request.getPrice());
        return mapToResponse(repository.save(p));
    }

    private ProductResponse mapToResponse(Product p) {
        return ProductResponse.builder().id(p.getId()).name(p.getName()).price(p.getPrice()).build();
    }
}

// --- Controller ---

@RestController
@RequestMapping("/api/products")
class ProductController {
    private final ProductService service;

    public ProductController(ProductService service) { this.service = service; }

    @GetMapping("/{id}")
    public ResponseEntity<ProductResponse> get(@PathVariable Long id) {
        return ResponseEntity.ok(service.getProduct(id));
    }

    @GetMapping
    public ResponseEntity<PagedResponse<ProductResponse>> search(
            ProductSearchCriteria criteria,
            @RequestParam(defaultValue = "0") @Min(0) int page,
            @RequestParam(defaultValue = "10") @Min(1) int size,
            @RequestParam(defaultValue = "id") String sort) {
        return ResponseEntity.ok(service.search(criteria, page, size, sort));
    }

    @PostMapping
    public ResponseEntity<ProductResponse> create(@Valid @RequestBody ProductRequest request) {
        return new ResponseEntity<>(service.create(request), HttpStatus.CREATED);
    }
}

// --- Exception Handling ---

@ControllerAdvice
class GlobalExceptionHandler {
    @ExceptionHandler(RuntimeException.class)
    public ResponseEntity<String> handleNotFound(RuntimeException ex) {
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(ex.getMessage());
    }
}