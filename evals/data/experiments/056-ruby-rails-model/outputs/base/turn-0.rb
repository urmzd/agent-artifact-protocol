class Order < ApplicationRecord
  belongs_to :user
  has_many :order_items, dependent: :destroy
  has_one :shipping_address, dependent: :destroy
  has_many :payments, dependent: :destroy

  enum status: { pending: 0, confirmed: 1, shipped: 2, delivered: 3, cancelled: 4 }

  validates :status, presence: true
  validates :total_amount, numericality: { greater_than_or_equal_to: 0 }
  validate :status_transition_valid, on: :update

  scope :recent, -> { order(created_at: :desc) }
  scope :by_status, ->(status) { where(status: status) }
  scope :by_date_range, ->(start_date, end_date) { where(created_at: start_date..end_date) }
  scope :high_value, ->(threshold = 1000) { where("total_amount > ?", threshold) }
  scope :pending_shipment, -> { where(status: [:confirmed]) }

  before_create :generate_order_number
  after_update :log_status_change, if: :saved_change_to_status?

  def subtotal
    order_items.sum(&:price)
  end

  def tax_amount
    subtotal * 0.08
  end

  def total
    subtotal + tax_amount
  end

  def apply_coupon(discount)
    self.total_amount -= discount
    save
  end

  def can_cancel?
    ['pending', 'confirmed'].include?(status)
  end

  def ship!
    return unless status == 'confirmed'
    update!(status: :shipped)
  end

  def complete!
    update!(status: :delivered)
  end

  private

  def generate_order_number
    self.order_number = "ORD-#{SecureRandom.hex(4).upcase}"
  end

  def status_transition_valid
    if status_changed?
      case status_change_to_hash['status']
      when ['shipped', 'cancelled']
        errors.add(:status, "cannot cancel shipped order")
      when ['delivered', 'cancelled']
        errors.add(:status, "cannot cancel delivered order")
      end
    end
  end

  def log_status_change
    Rails.logger.info "Order #{id} changed from #{saved_changes[:status][0]} to #{saved_changes[:status][1]}"
  end
end